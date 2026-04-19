// Package main — adversarial_review and gate_review tool implementations.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed embedded/references/*.md
var reviewReferencesFS embed.FS

// bofPool is the 3-slot model pool for adversarial review rotation.
// Pool index = state.Iteration % 3.
var bofPool = []string{
	"copilot/gpt-4.1",        // slot 0
	"copilot/claude-opus-4",  // slot 1
	"copilot/gpt-4o",         // slot 2
}

// validModelRe matches valid exclude_model values: alphanumeric, hyphen, underscore, dot, slash.
var validModelRe = regexp.MustCompile(`^[a-zA-Z0-9_./-]+$`)

// verdictRe parses the Verdict line from review output.
var verdictRe = regexp.MustCompile(`(?m)^Verdict:\s*(PASSED|CONDITIONAL|FAILED)`)

// mcpErr returns an MCP error result without propagating a Go error.
func mcpErr(format string, args ...any) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(format, args...)},
		},
	}, nil, nil
}

// extractVerdict returns the verdict string from output.
func extractVerdict(output string) string {
	m := verdictRe.FindStringSubmatch(output)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// loadReferenceContent reads all embedded reference .md files and
// concatenates them with headers for use in the review preamble.
func loadReferenceContent() string {
	var sb strings.Builder
	_ = fs.WalkDir(reviewReferencesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, readErr := reviewReferencesFS.ReadFile(path)
		if readErr != nil {
			return nil
		}
		sb.WriteString("\n--- ")
		sb.WriteString(filepath.Base(path))
		sb.WriteString(" ---\n")
		sb.Write(data)
		sb.WriteString("\n")
		return nil
	})
	return sb.String()
}

// excludeModelFromPool returns a copy of pool with matching entries removed.
// If exclusion would empty the pool, the original pool is returned unchanged.
func excludeModelFromPool(pool []string, exclude string) []string {
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
	return filtered
}

// adversarialInput is the input schema for the adversarial_review tool.
type adversarialInput struct {
	PlanSlug     string `json:"plan_slug"               jsonschema:"Plan slug used as state file name"`
	PlanContent  string `json:"plan_content"            jsonschema:"Full text of the plan to review"`
	ExcludeModel string `json:"exclude_model,omitempty" jsonschema:"Full model ID to exclude from review pool (e.g. copilot/gpt-4.1). Empty or omitted = no exclusion."`
}

// newAdversarialHandler returns the adversarial_review MCP tool handler.
func newAdversarialHandler(projectRoot string) func(context.Context, *mcp.CallToolRequest, adversarialInput) (*mcp.CallToolResult, any, error) {
	references := loadReferenceContent()
	return func(ctx context.Context, req *mcp.CallToolRequest, input adversarialInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.PlanContent) == "" {
			return mcpErr("plan_content must not be empty")
		}
		if strings.TrimSpace(input.PlanSlug) == "" {
			return mcpErr("plan_slug must not be empty")
		}

		pool := excludeModelFromPool(bofPool, input.ExcludeModel)

		state, err := ReadState(projectRoot, input.PlanSlug)
		if err != nil {
			return mcpErr("failed to read state: %v", err)
		}

		// Simple modulo rotation: compute slot from bofPool (always 3 elements)
		// so the rotation is stable across calls. If the computed slot was filtered
		// out by exclude_model, the filtered pool still has ≥1 entry (guarded by
		// excludeModelFromPool), so we take modulo over the filtered pool as fallback.
		slot := state.Iteration % 3
		selectedModel := bofPool[slot]
		// Check if selectedModel survived the filter; if not, fall back to pool[slot%len(pool)].
		{
			found := false
			for _, m := range pool {
				if m == selectedModel {
					found = true
					break
				}
			}
			if !found {
				selectedModel = pool[slot%len(pool)]
			}
		}

		rctx, cancel := context.WithTimeout(ctx, 300*time.Second)
		defer cancel()

		date := time.Now().UTC().Format("2006-01-02")
		preamble := fmt.Sprintf(
			"You are adversarial reviewer (iteration %d) for the bof adversarial review workflow.\n"+
				"Apply the 7-attack protocol described in the references below.\n"+
				"Project root: %s\n"+
				"Write your review report to %s/.adversarial/reports/review-%s-iter%d-%s.md\n"+
				"Do NOT write the state file — the handler writes it after review completes.\n"+
				"The final line of your report MUST be: Verdict: PASSED|CONDITIONAL|FAILED\n\n"+
				"=== REVIEW REFERENCES ===\n%s\n"+
				"=== PLAN TO REVIEW ===\n",
			state.Iteration,
			projectRoot,
			projectRoot, date, state.Iteration, input.PlanSlug,
			references,
		)

		prompt := preamble + input.PlanContent
		res, runErr := runCrushFn(rctx, selectedModel, prompt)
		if runErr != nil {
			return mcpErr("review round failed: %v", runErr)
		}

		verdict := extractVerdict(res.Output)
		if verdict == "" {
			log.Printf("bof-mcp: adversarial review produced no valid Verdict: line")
		}

		state.Iteration++
		state.LastModel = selectedModel
		state.LastVerdict = verdict
		state.LastReviewDate = date
		if err := WriteState(projectRoot, state); err != nil {
			return mcpErr("failed to write state: %v", err)
		}

		summary := fmt.Sprintf("\n\n=== Summary ===\nModel: %s\nVerdict: %s\n", selectedModel, verdict)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: res.Output + summary}},
		}, nil, nil
	}
}

// gateInput is the input schema for the gate_review tool.
type gateInput struct {
	Strict bool `json:"strict" jsonschema:"If true, block when no state files exist"`
}

// gateOutput is the structured response for gate_review.
type gateOutput struct {
	Blocked       bool     `json:"blocked"`
	Reason        string   `json:"reason"`
	BlockingPlans []string `json:"blocking_plans,omitempty"`
}

// newGateHandler returns the gate_review MCP tool handler.
func newGateHandler(projectRoot string) func(context.Context, *mcp.CallToolRequest, gateInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input gateInput) (*mcp.CallToolResult, any, error) {
		rawDir := stateDir(projectRoot)
		dir, err := filepath.Abs(filepath.Clean(rawDir))
		if err != nil {
			dir = filepath.Clean(rawDir)
		}
		entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			entries = nil
		}
		// Filter to files directly in dir (not in subdirectories like reports/).
		var files []string
		for _, e := range entries {
			if filepath.Dir(e) == dir {
				files = append(files, e)
			}
		}

		if len(files) == 0 {
			if input.Strict {
				return gateResult(gateOutput{
					Blocked: true,
					Reason:  "adversarial review required before completing this session",
				})
			}
			return gateResult(gateOutput{
				Blocked: false,
				Reason:  "no reviews in progress",
			})
		}

		var blocking []string
		for _, f := range files {
			data, readErr := os.ReadFile(f)
			if readErr != nil {
				continue
			}
			var s ReviewState
			if unmErr := json.Unmarshal(data, &s); unmErr != nil {
				continue
			}
			v := strings.ToUpper(strings.TrimSpace(s.LastVerdict))
			if v != "PASSED" && v != "CONDITIONAL" {
				slug := strings.TrimSuffix(filepath.Base(f), ".json")
				blocking = append(blocking, slug)
			}
		}

		if len(blocking) > 0 {
			return gateResult(gateOutput{
				Blocked:       true,
				Reason:        fmt.Sprintf("%d plan(s) have FAILED or missing verdicts", len(blocking)),
				BlockingPlans: blocking,
			})
		}
		return gateResult(gateOutput{
			Blocked: false,
			Reason:  "all plans have PASSED or CONDITIONAL verdicts",
		})
	}
}

func gateResult(out gateOutput) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "internal error: " + err.Error()}},
		}, nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}
