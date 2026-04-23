// Package main — model pool management for adversarial review rotation.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	defaultRounds = 5
	maxRounds     = 50
)

// defaultModels is the default 5-slot model pool.
var defaultModels = []string{
	"copilot/claude-sonnet-4.6",
	"copilot/gpt-4.1",
	"copilot/claude-opus-4",
	"copilot/gpt-4o",
	"gemini/gemini-2.5-pro-preview-05-06",
}

// validModelRe matches valid model IDs: alphanumeric, hyphen, underscore, dot, slash.
var validModelRe = regexp.MustCompile(`^[a-zA-Z0-9_./-]+$`)

// runCrushFn is the function used to invoke crush — replaceable in tests.
var runCrushFn = RunCrush

// randSource is used by buildRotationOrder; replaceable via SetRandSource.
var randSource rand.Source

// SetRandSource replaces the random source used by buildRotationOrder.
// Not goroutine-safe — intended for test use only.
func SetRandSource(src rand.Source) {
	randSource = src
}

// newRand returns a *rand.Rand seeded from randSource if set, otherwise from time.Now().UnixNano().
func newRand() *rand.Rand {
	if randSource != nil {
		return rand.New(randSource)
	}
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// effectiveRounds clamps n to [1, maxRounds], defaulting to defaultRounds when n < 1.
func effectiveRounds(n int) int {
	if n < 1 {
		return defaultRounds
	}
	if n > maxRounds {
		log.Printf("bof-mcp: rounds=%d exceeds maxRounds=%d, clamping", n, maxRounds)
		return maxRounds
	}
	return n
}

// buildModelPool returns the model pool from BOF_MODELS env var.
// Falls back to defaultModels if the var is unset or all entries are invalid.
func buildModelPool() []string {
	raw := os.Getenv("BOF_MODELS")
	if raw == "" {
		return append([]string(nil), defaultModels...)
	}
	var pool []string
	for _, m := range strings.Split(raw, ",") {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if !validModelRe.MatchString(m) {
			log.Printf("bof-mcp: BOF_MODELS entry %q contains invalid characters, skipping", m)
			continue
		}
		parts := strings.SplitN(m, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			log.Printf("bof-mcp: BOF_MODELS entry %q must be provider/model, skipping", m)
			continue
		}
		pool = append(pool, m)
	}
	if len(pool) == 0 {
		log.Printf("bof-mcp: BOF_MODELS produced no valid entries, falling back to defaults")
		return append([]string(nil), defaultModels...)
	}
	return pool
}

// excludeModelFilter returns a copy of pool with all entries that exactly match
// exclude (case-insensitive) removed.
// If exclude is empty: no-op. If exclusion would empty the pool: fail-open (returns pool unchanged).
func excludeModelFilter(pool []string, exclude string) []string {
	exclude = strings.TrimSpace(exclude)
	if exclude == "" {
		return pool
	}
	log.Printf("bof-mcp: exclude_model=%q requested", exclude)
	if !validModelRe.MatchString(exclude) {
		log.Printf("bof-mcp: exclude_model=%q is malformed, ignoring", exclude)
		return pool
	}
	excludeLower := strings.ToLower(exclude)
	filtered := make([]string, 0, len(pool))
	for _, m := range pool {
		if strings.ToLower(m) != excludeLower {
			filtered = append(filtered, m)
		}
	}
	if len(filtered) == 0 {
		log.Printf("bof-mcp: exclude_model=%q would empty pool; ignoring exclusion", exclude)
		return pool
	}
	if len(filtered) == len(pool) {
		log.Printf("bof-mcp: exclude_model=%q matched no pool entries (no-op)", exclude)
	}
	return filtered
}

// errAllModelsUnavailable is returned when every model in the pool is blocked.
var errAllModelsUnavailable = errors.New("all models in the pool are unavailable " +
	"(enterprise policy may be blocking them); " +
	"set BOF_MODELS to a list of models known to be accessible")

// modelUnavailablePatterns are case-insensitive substrings that indicate a model
// is blocked by enterprise policy rather than a transient error.
var modelUnavailablePatterns = []string{
	"model is not supported",
	"model not supported",
	"not available for your organization",
	"not enabled for your organization",
	"access to this model",
	"model access denied",
	"this model is not available",
	"not supported via",
}

// providerOf extracts the provider prefix (before the first "/") from a model string.
func providerOf(model string) string {
	if idx := strings.Index(model, "/"); idx >= 0 {
		return model[:idx]
	}
	return model
}

// familyInterleaveShuffle returns a permutation of pool where models from the
// same provider are spread out as evenly as possible.
// Uses rng for intra-group shuffling; O(n²), acceptable for n≤10.
func familyInterleaveShuffle(pool []string, rng *rand.Rand) []string {
	order := make([]string, 0, len(pool))
	groups := make(map[string][]string)
	providers := make([]string, 0)
	for _, m := range pool {
		p := providerOf(m)
		if _, seen := groups[p]; !seen {
			providers = append(providers, p)
		}
		groups[p] = append(groups[p], m)
	}
	for _, p := range providers {
		g := groups[p]
		rng.Shuffle(len(g), func(i, j int) { g[i], g[j] = g[j], g[i] })
		groups[p] = g
	}
	sort.Strings(providers)
	remaining := make(map[string][]string, len(groups))
	for k, v := range groups {
		remaining[k] = append([]string(nil), v...)
	}
	last := ""
	for len(order) < len(pool) {
		best := ""
		bestCount := -1
		for _, p := range providers {
			if p == last {
				continue
			}
			if cnt := len(remaining[p]); cnt > bestCount {
				bestCount = cnt
				best = p
			}
		}
		if best == "" || bestCount == 0 {
			best = last
		}
		order = append(order, remaining[best][0])
		remaining[best] = remaining[best][1:]
		if len(remaining[best]) == 0 {
			delete(remaining, best)
		}
		last = best
	}
	return order
}

// buildRotationOrder returns a slice of model strings of length rounds, drawn
// from pool in family-interleaved batches of batchSize (5).
func buildRotationOrder(pool []string, rounds int) []string {
	const batchSize = 5
	rng := newRand()
	result := make([]string, 0, rounds)
	for len(result) < rounds {
		batch := familyInterleaveShuffle(pool, rng)
		for len(batch) < batchSize {
			extra := familyInterleaveShuffle(pool, rng)
			if len(batch) > 0 && len(extra) > 0 && batch[len(batch)-1] == extra[0] {
				if len(extra) > 1 {
					extra[0], extra[1] = extra[1], extra[0]
				}
			}
			batch = append(batch, extra...)
		}
		if len(result) > 0 && batch[0] == result[len(result)-1] {
			if len(batch) > 1 {
				batch[0], batch[1] = batch[1], batch[0]
			}
		}
		take := batchSize
		if rounds-len(result) < take {
			take = rounds - len(result)
		}
		result = append(result, batch[:take]...)
	}
	return result
}

// worstVerdict returns the most severe verdict from the slice.
// Severity order: FAILED > CONDITIONAL > PASSED > "".
func worstVerdict(verdicts []string) string {
	worst := ""
	for _, v := range verdicts {
		switch v {
		case "FAILED":
			return "FAILED"
		case "CONDITIONAL":
			if worst != "FAILED" {
				worst = "CONDITIONAL"
			}
		case "PASSED":
			if worst == "" {
				worst = "PASSED"
			}
		}
	}
	return worst
}

// isModelUnavailable reports whether the crush output indicates the model is
// blocked by enterprise policy rather than a transient error.
func isModelUnavailable(output string) bool {
	lower := strings.ToLower(output)
	for _, pattern := range modelUnavailablePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// runOneRound runs the adversarial review prompt against targetModel, falling
// back to other pool models if the primary is unavailable.
// bof-mcp uses stdin-based prompt passing (strings passed directly, no temp files).
func runOneRound(ctx context.Context, pool []string, targetModel, preamble, planContent string) (usedModel, output string, err error) {
	prompt := preamble + "\n--- PLAN CONTENT ---\n" + planContent

	tryModel := func(model string) (string, bool, error) {
		res, runErr := runCrushFn(ctx, model, prompt)
		if runErr != nil {
			return "", false, runErr
		}
		if res.ExitCode == 0 {
			return res.Output, false, nil
		}
		if isModelUnavailable(res.Output) {
			return res.Output, true, nil
		}
		return res.Output, false, fmt.Errorf("crush exited %d: %s", res.ExitCode, res.Output)
	}

	out, unavailable, runErr := tryModel(targetModel)
	if runErr != nil {
		return "", out, runErr
	}
	if !unavailable {
		return targetModel, out, nil
	}

	// Primary unavailable — try remaining pool models.
	for _, m := range pool {
		if m == targetModel {
			continue
		}
		out, unavailable, runErr = tryModel(m)
		if runErr != nil {
			return "", out, runErr
		}
		if !unavailable {
			return m, out, nil
		}
	}

	return "", "", errAllModelsUnavailable
}