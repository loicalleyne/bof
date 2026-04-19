// Package main — implementer_agent, spec_review, quality_review tool implementations.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed embedded/agents/ImplementerAgent.agent.md
var implementerAgentMD []byte

//go:embed embedded/agents/SpecReviewerAgent.agent.md
var specReviewerAgentMD []byte

//go:embed embedded/agents/CodeQualityReviewerAgent.agent.md
var codeQualityReviewerAgentMD []byte

// stripFrontmatter removes YAML frontmatter from markdown content.
// Finds the first "---\n", finds the closing "---\n" after it, and discards
// both delimiters and everything between them. If no closing delimiter is found,
// the full content is returned unchanged.
func stripFrontmatter(content string) string {
	const delim = "---\n"
	start := strings.Index(content, delim)
	if start == -1 {
		return content
	}
	rest := content[start+len(delim):]
	end := strings.Index(rest, delim)
	if end == -1 {
		return content
	}
	return rest[end+len(delim):]
}

// implementerInput is the input schema for the implementer_agent tool.
type implementerInput struct {
	TaskContent  string `json:"task_content"            jsonschema:"Full text of the task document to implement"`
	Model        string `json:"model,omitempty"         jsonschema:"Model ID to use (e.g. copilot/claude-sonnet-4.6). Falls back to --default-model if omitted."`
	ProjectRoot  string `json:"project_root,omitempty"  jsonschema:"Override project root for this invocation. Defaults to server --project-root."`
}

// specReviewInput is the input schema for the spec_review tool.
type specReviewInput struct {
	SpecContent string `json:"spec_content"            jsonschema:"Full text of the spec or design document to review"`
	Model       string `json:"model,omitempty"         jsonschema:"Model ID to use. Falls back to --default-model if omitted."`
}

// qualityReviewInput is the input schema for the quality_review tool.
type qualityReviewInput struct {
	CodeContent string `json:"code_content"            jsonschema:"Full text of code or diff to review for quality"`
	Model       string `json:"model,omitempty"         jsonschema:"Model ID to use. Falls back to --default-model if omitted."`
}

// newImplementerHandler returns the implementer_agent MCP tool handler.
func newImplementerHandler(serverProjectRoot, defaultModel string) func(context.Context, *mcp.CallToolRequest, implementerInput) (*mcp.CallToolResult, any, error) {
	agentInstructions := stripFrontmatter(string(implementerAgentMD))
	return func(ctx context.Context, req *mcp.CallToolRequest, input implementerInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.TaskContent) == "" {
			return mcpErr("task_content must not be empty")
		}
		model := resolveModel(input.Model, defaultModel)
		if model == "" {
			return mcpErr("no model specified and --default-model not set; provide model in request or start server with --default-model")
		}

		root := serverProjectRoot
		if strings.TrimSpace(input.ProjectRoot) != "" {
			root = strings.TrimSpace(input.ProjectRoot)
		}

		preamble := fmt.Sprintf(
			"You are an ImplementerAgent. Your project root is: %s\n\n"+
				"=== IMPLEMENTER AGENT INSTRUCTIONS ===\n%s\n"+
				"=== TASK DOCUMENT ===\n",
			root,
			agentInstructions,
		)
		prompt := preamble + input.TaskContent

		res, err := runCrushFn(ctx, model, prompt)
		if err != nil {
			return mcpErr("implementer_agent dispatch failed: %v", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: res.Output}},
		}, nil, nil
	}
}

// newSpecReviewHandler returns the spec_review MCP tool handler.
func newSpecReviewHandler(defaultModel string) func(context.Context, *mcp.CallToolRequest, specReviewInput) (*mcp.CallToolResult, any, error) {
	agentInstructions := stripFrontmatter(string(specReviewerAgentMD))
	return func(ctx context.Context, req *mcp.CallToolRequest, input specReviewInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.SpecContent) == "" {
			return mcpErr("spec_content must not be empty")
		}
		model := resolveModel(input.Model, defaultModel)
		if model == "" {
			return mcpErr("no model specified and --default-model not set; provide model in request or start server with --default-model")
		}

		preamble := "You are a SpecReviewerAgent.\n\n" +
			"=== SPEC REVIEWER INSTRUCTIONS ===\n" + agentInstructions +
			"\n=== SPEC DOCUMENT ===\n"
		prompt := preamble + input.SpecContent

		res, err := runCrushFn(ctx, model, prompt)
		if err != nil {
			return mcpErr("spec_review dispatch failed: %v", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: res.Output}},
		}, nil, nil
	}
}

// newQualityReviewHandler returns the quality_review MCP tool handler.
func newQualityReviewHandler(defaultModel string) func(context.Context, *mcp.CallToolRequest, qualityReviewInput) (*mcp.CallToolResult, any, error) {
	agentInstructions := stripFrontmatter(string(codeQualityReviewerAgentMD))
	return func(ctx context.Context, req *mcp.CallToolRequest, input qualityReviewInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.CodeContent) == "" {
			return mcpErr("code_content must not be empty")
		}
		model := resolveModel(input.Model, defaultModel)
		if model == "" {
			return mcpErr("no model specified and --default-model not set; provide model in request or start server with --default-model")
		}

		preamble := "You are a CodeQualityReviewerAgent.\n\n" +
			"=== CODE QUALITY REVIEWER INSTRUCTIONS ===\n" + agentInstructions +
			"\n=== CODE TO REVIEW ===\n"
		prompt := preamble + input.CodeContent

		res, err := runCrushFn(ctx, model, prompt)
		if err != nil {
			return mcpErr("quality_review dispatch failed: %v", err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: res.Output}},
		}, nil, nil
	}
}

// resolveModel returns the first non-empty value among provided, fallback.
func resolveModel(provided, fallback string) string {
	if strings.TrimSpace(provided) != "" {
		return strings.TrimSpace(provided)
	}
	return strings.TrimSpace(fallback)
}
