// Package main — model discovery and background probe management.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// validModelIDRe matches valid model IDs: alphanumeric, hyphen, underscore, dot, slash.
var validModelIDRe = regexp.MustCompile(`^[a-zA-Z0-9_./-]+$`)

// runCrushFn is the function used to invoke crush — replaceable in tests.
var runCrushFn = RunCrush

// ModelEntry represents one probed model in the availability cache.
type ModelEntry struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Available bool      `json:"available"`
	ProbedAt  time.Time `json:"probed_at"`
}

// ModelCache is the JSON schema for ~/.config/bof/model-cache.json.
type ModelCache struct {
	Entries        []ModelEntry `json:"entries"`
	CachedAt       time.Time    `json:"cached_at"`
	ProbeCompleted bool         `json:"probe_completed"`
}

// modelProber manages the background probe goroutine and disk cache.
// All shared state is protected by mu.
// RLock for reads; Lock only during probe state transitions and cache writes.
type modelProber struct {
	mu           sync.RWMutex
	cache        *ModelCache
	probing      bool
	probeErrors  []string
	cacheErr     string
	cancelProbe  context.CancelFunc
	done         chan struct{} // closed when current probe completes
	cachePath    string
	ttl          time.Duration
	// Injectable for testing.
	listModelsFn func(ctx context.Context) ([]string, error)
	probeFn      func(ctx context.Context, model string) bool
}

// defaultCachePath returns ~/.config/bof/model-cache.json.
func defaultCachePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "bof", "model-cache.json"), nil
}

// defaultProbeTTL returns the probe TTL, defaulting to 1 day.
// Override with BOF_MODEL_CACHE_TTL_DAYS env var.
func defaultProbeTTL() time.Duration {
	envVal := os.Getenv("BOF_MODEL_CACHE_TTL_DAYS")
	if envVal != "" {
		days, err := strconv.Atoi(envVal)
		if err != nil || days <= 0 {
			log.Printf("bof-mcp: invalid BOF_MODEL_CACHE_TTL_DAYS=%q, using default (1)", envVal)
		} else {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	return 24 * time.Hour
}

// providerOf extracts the provider prefix (before the first "/") from a model string.
func providerOf(model string) string {
	if idx := strings.Index(model, "/"); idx >= 0 {
		return model[:idx]
	}
	return model
}

// newModelProber creates a modelProber with default crush-based list and probe functions.
func newModelProber(cachePath string, ttl time.Duration) *modelProber {
	return newModelProberWithFuncs(cachePath, ttl,
		func(ctx context.Context) ([]string, error) {
			crushPath, err := exec.LookPath("crush")
			if err != nil {
				return nil, fmt.Errorf("crush not in PATH: %w", err)
			}
			cmd := exec.CommandContext(ctx, crushPath, "models")
			out, err := cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("crush models failed: %w", err)
			}
			var models []string
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if !validModelIDRe.MatchString(line) {
					log.Printf("bof-mcp: skipping invalid model ID from crush models: %q", line)
					continue
				}
				models = append(models, line)
			}
			return models, nil
		},
		func(ctx context.Context, model string) bool {
			// Probe with "1" as prompt content; 15s timeout enforced by caller.
			res, err := runCrushFn(ctx, model, "1")
			if err != nil {
				return false
			}
			if res.ExitCode == 0 {
				return true
			}
			lower := strings.ToLower(res.Output)
			for _, pattern := range []string{
				"model is not supported",
				"model not supported",
				"not available for your organization",
				"not enabled for your organization",
				"access to this model",
				"model access denied",
				"this model is not available",
			} {
				if strings.Contains(lower, pattern) {
					return false
				}
			}
			return true // transient error: fail-open
		},
	)
}

// newModelProberWithFuncs creates a modelProber with injectable functions (for testing).
func newModelProberWithFuncs(cachePath string, ttl time.Duration, listFn func(context.Context) ([]string, error), probeFn func(context.Context, string) bool) *modelProber {
	p := &modelProber{
		cachePath:    cachePath,
		ttl:          ttl,
		listModelsFn: listFn,
		probeFn:      probeFn,
	}
	if cachePath == "" {
		log.Printf("bof-mcp: running without model-cache.json (UserConfigDir failed)")
	} else {
		if err := p.loadCache(); err != nil && !os.IsNotExist(err) {
			log.Printf("bof-mcp: failed to load model cache: %v", err)
		}
	}
	return p
}

func (p *modelProber) loadCache() error {
	if p.cachePath == "" {
		return nil
	}
	data, err := os.ReadFile(p.cachePath)
	if err != nil {
		return err
	}
	var cache ModelCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return err
	}
	// Drop invalid model IDs from cached entries.
	valid := make([]ModelEntry, 0, len(cache.Entries))
	for _, e := range cache.Entries {
		if validModelIDRe.MatchString(e.ID) {
			valid = append(valid, e)
		} else {
			log.Printf("bof-mcp: invalid model ID in cache: %q, dropping", e.ID)
		}
	}
	cache.Entries = valid

	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache = &cache
	return nil
}

// saveCache atomically writes the cache to disk via temp file + rename.
// On failure it stores the error string in p.cacheErr (returned by discover_models).
func (p *modelProber) saveCache() {
	if p.cachePath == "" {
		return
	}

	p.mu.RLock()
	cache := p.cache
	p.mu.RUnlock()
	if cache == nil {
		return
	}

	data, err := json.Marshal(cache)
	if err != nil {
		log.Printf("bof-mcp: failed to encode model cache: %v", err)
		p.mu.Lock()
		p.cacheErr = fmt.Sprintf("encode cache: %v", err)
		p.mu.Unlock()
		return
	}

	dir := filepath.Dir(p.cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Printf("bof-mcp: failed to create cache dir %q: %v", dir, err)
		p.mu.Lock()
		p.cacheErr = fmt.Sprintf("mkdir cache dir: %v", err)
		p.mu.Unlock()
		return
	}

	tmp, err := os.CreateTemp(dir, "model-cache-*.json")
	if err != nil {
		log.Printf("bof-mcp: failed to create cache temp file: %v", err)
		p.mu.Lock()
		p.cacheErr = fmt.Sprintf("create temp file: %v", err)
		p.mu.Unlock()
		return
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		p.mu.Lock()
		p.cacheErr = fmt.Sprintf("write temp file: %v", err)
		p.mu.Unlock()
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		p.mu.Lock()
		p.cacheErr = fmt.Sprintf("close temp file: %v", err)
		p.mu.Unlock()
		return
	}
	if err := os.Rename(tmpName, p.cachePath); err != nil {
		_ = os.Remove(tmpName)
		log.Printf("bof-mcp: failed to rename cache temp file: %v", err)
		p.mu.Lock()
		p.cacheErr = fmt.Sprintf("rename cache: %v", err)
		p.mu.Unlock()
		return
	}
	p.mu.Lock()
	p.cacheErr = "" // clear any prior error on success
	p.mu.Unlock()
}

// startProbe launches the background probe goroutine if not already running.
// Non-blocking: returns immediately after launching the goroutine.
func (p *modelProber) startProbe(ctx context.Context) {
	p.mu.Lock()
	if p.probing {
		p.mu.Unlock()
		return
	}
	p.probing = true
	p.probeErrors = nil
	p.done = make(chan struct{})
	probeCtx, cancel := context.WithCancel(ctx)
	p.cancelProbe = cancel
	done := p.done
	p.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("bof-mcp: model probe panic: %v", r)
			}
			p.mu.Lock()
			p.probing = false
			close(done)
			p.mu.Unlock()
			cancel()
		}()

		listCtx, cancelList := context.WithTimeout(probeCtx, 10*time.Second)
		defer cancelList()
		models, err := p.listModelsFn(listCtx)
		if err != nil {
			log.Printf("bof-mcp: failed to list models: %v", err)
			p.mu.Lock()
			p.probeErrors = []string{fmt.Sprintf("list models: %v", err)}
			p.mu.Unlock()
			return
		}

		now := time.Now().UTC()
		entries := make([]ModelEntry, len(models))
		for i, m := range models {
			entries[i] = ModelEntry{
				ID:        m,
				Provider:  providerOf(m),
				Available: true,
				ProbedAt:  now,
			}
		}

		p.mu.Lock()
		p.cache = &ModelCache{Entries: entries, CachedAt: now}
		p.mu.Unlock()

		var probeErrs []string
		for i, m := range models {
			select {
			case <-probeCtx.Done():
				return
			default:
			}

			modelCtx, cancelModel := context.WithTimeout(probeCtx, 15*time.Second)
			avail := p.probeFn(modelCtx, m)
			cancelModel()

			p.mu.Lock()
			if p.cache != nil && i < len(p.cache.Entries) {
				p.cache.Entries[i].Available = avail
				p.cache.Entries[i].ProbedAt = time.Now().UTC()
			}
			p.mu.Unlock()

			if !avail {
				probeErrs = append(probeErrs, fmt.Sprintf("model %s: unavailable or probe failed", m))
			}
		}

		p.mu.Lock()
		if p.cache != nil {
			p.cache.ProbeCompleted = true
			p.cache.CachedAt = time.Now().UTC()
		}
		if len(probeErrs) == len(models) && len(models) > 0 {
			p.probeErrors = probeErrs
		}
		p.mu.Unlock()

		p.saveCache()
	}()
}

// startProbeIfStale starts a probe only if the cache is missing or stale.
func (p *modelProber) startProbeIfStale(ctx context.Context) {
	p.mu.RLock()
	stale := p.cache == nil || time.Since(p.cache.CachedAt) > p.ttl
	probing := p.probing
	p.mu.RUnlock()

	if probing || !stale {
		return
	}
	p.startProbe(ctx)
}

// forceRefresh cancels any in-flight probe, waits for it to finish, then
// starts a new probe.
func (p *modelProber) forceRefresh(ctx context.Context) {
	p.mu.Lock()
	if p.probing && p.cancelProbe != nil {
		p.cancelProbe()
	}
	p.mu.Unlock()

	// Wait for old probe to finish (context cancellation kills crush subprocess quickly).
	p.mu.RLock()
	d := p.done
	p.mu.RUnlock()
	if d != nil {
		<-d
	}

	p.startProbe(ctx)
}

// currentState returns a snapshot of the current cache state.
// Uses RLock for concurrent-read safety.
func (p *modelProber) currentState() (entries []ModelEntry, probing bool, stale bool, cachedAt time.Time, probeErrors []string, cacheErr string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cache != nil {
		// Copy the slice to avoid a data race: the probe goroutine modifies
		// individual elements of p.cache.Entries while holding Lock; callers
		// iterate the returned slice after RUnlock with no lock held.
		entries = make([]ModelEntry, len(p.cache.Entries))
		copy(entries, p.cache.Entries)
		cachedAt = p.cache.CachedAt
		stale = time.Since(cachedAt) > p.ttl
	}
	probing = p.probing
	probeErrors = p.probeErrors
	cacheErr = p.cacheErr
	return
}

// discoverInput is the input schema for the discover_models tool.
type discoverInput struct {
	Filter       string `json:"filter,omitempty"        jsonschema:"Optional substring filter for model names (max 200 chars)"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Set to true to clear the cache and trigger a new background probe"`
}

// newDiscoverHandler returns an MCP handler that lists available crush models.
func newDiscoverHandler(prober *modelProber) func(context.Context, *mcp.CallToolRequest, discoverInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input discoverInput) (*mcp.CallToolResult, any, error) {
		if len(input.Filter) > 200 {
			return mcpErr("filter must not exceed 200 characters")
		}

		if input.ForceRefresh {
			prober.forceRefresh(ctx)
		} else {
			_, probing, _, _, _, _ := prober.currentState()
			if !probing {
				prober.startProbeIfStale(ctx)
			}
		}

		entries, probing, stale, cachedAt, probeErrors, cacheErr := prober.currentState()

		var models []ModelEntry
		lowerFilter := strings.ToLower(strings.TrimSpace(input.Filter))
		for _, m := range entries {
			if lowerFilter != "" && !strings.Contains(strings.ToLower(m.ID), lowerFilter) {
				continue
			}
			models = append(models, m)
		}
		if models == nil {
			models = []ModelEntry{} // ensure JSON array not null
		}

		resp := map[string]interface{}{
			"models":  models,
			"probing": probing,
		}
		if stale {
			resp["stale"] = true
		}
		if !cachedAt.IsZero() {
			resp["cached_at"] = cachedAt.Format(time.RFC3339)
		}
		if len(probeErrors) > 0 {
			resp["probe_errors"] = probeErrors
		}
		if cacheErr != "" {
			resp["cache_error"] = cacheErr
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			return mcpErr("failed to encode discover_models response: %v", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(respBytes)}},
		}, nil, nil
	}
}
