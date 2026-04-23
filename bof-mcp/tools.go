// Package main — tool registration.
package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerTools registers all bof-mcp tools on the given server.
// When noAdversarial is true, adversarial_review and gate_review are omitted.
func registerTools(server *mcp.Server, projectRoot, defaultModel string, noAdversarial bool) {
	if !noAdversarial {
		mcp.AddTool(server, &mcp.Tool{
			Name: "adversarial_review",
			Description: "Dispatch one or more adversarial review rounds for the given plan using a " +
				"family-interleaved model pool (default: 5 models from copilot, gemini). " +
				"Reads .adversarial/{plan_slug}.json for iteration state, runs the requested rounds, " +
				"and writes the worst verdict back to the state file. " +
				"Configure the pool with BOF_MODELS (comma-separated provider/model). " +
				"Optional: pass rounds (default 5, max 50) and exclude_model to skip the caller's model.",
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

}
