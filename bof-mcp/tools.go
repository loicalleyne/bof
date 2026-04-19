// Package main — tool registration.
package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerTools registers all bof-mcp tools on the given server.
// When noAdversarial is true, adversarial_review and gate_review are omitted.
func registerTools(server *mcp.Server, projectRoot, defaultModel string, noAdversarial bool, prober *modelProber) {
	if !noAdversarial {
		mcp.AddTool(server, &mcp.Tool{
			Name: "adversarial_review",
			Description: "Dispatch one adversarial review round for the given plan using a 3-slot " +
				"rotation pool (copilot/gpt-4.1, copilot/claude-opus-4, copilot/gpt-4o). " +
				"Reads .adversarial/{plan_slug}.json for iteration state, runs one review pass, " +
				"and writes the verdict back to the state file. " +
				"Pool rotation uses simple modulo: pool[iteration % 3]. " +
				"Optional: pass exclude_model to skip the current agent's model.",
		}, newAdversarialHandler(projectRoot))

		mcp.AddTool(server, &mcp.Tool{
			Name:        "gate_review",
			Description: "Check whether all adversarial review verdicts in .adversarial/ are PASSED or CONDITIONAL.",
		}, newGateHandler(projectRoot))
	}

	mcp.AddTool(server, &mcp.Tool{
		Name: "implementer_agent",
		Description: "Dispatch an ImplementerAgent to implement a task document completely, following TDD. " +
			"Pass the full task document text as task_content. " +
			"The agent follows the bof ImplementerAgent protocol (startup, TDD cycle, completion). " +
			"Specify model to override the server default.",
	}, newImplementerHandler(projectRoot, defaultModel))

	mcp.AddTool(server, &mcp.Tool{
		Name: "spec_review",
		Description: "Dispatch a SpecReviewerAgent to review a specification or design document. " +
			"Pass the full spec text as spec_content. " +
			"The agent produces a structured review with findings and recommendations. " +
			"Specify model to override the server default.",
	}, newSpecReviewHandler(defaultModel))

	mcp.AddTool(server, &mcp.Tool{
		Name: "quality_review",
		Description: "Dispatch a CodeQualityReviewerAgent to review code or a diff for quality issues. " +
			"Pass the full code or diff as code_content. " +
			"The agent checks for correctness, style, security, and performance concerns. " +
			"Specify model to override the server default.",
	}, newQualityReviewHandler(defaultModel))

	mcp.AddTool(server, &mcp.Tool{
		Name: "discover_models",
		Description: "List available crush models and their probe status. " +
			"Returns a JSON object with models array, probing flag, and optional stale/cached_at/probe_errors/cache_error fields. " +
			"Call this first to see which models are available before using dispatch tools. " +
			"Set force_refresh=true to trigger a new background probe immediately. " +
			"TTL is configured via BOF_MODEL_CACHE_TTL_DAYS (default: 1 day).",
	}, newDiscoverHandler(prober))
}
