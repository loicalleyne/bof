// Command bof-mcp is a lightweight MCP stdio server that exposes agent dispatch
// and adversarial review tools for the bof framework, compatible with Crush and
// VS Code via MCP.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if runtime.GOOS == "windows" {
		fmt.Fprintln(os.Stderr, "bof-mcp does not support Windows — run from Linux, macOS, or WSL")
		os.Exit(1)
	}

	// 1. Parse flags.
	projectRootFlag := flag.String("project-root", "", "project root directory (default: $BOF_PROJECT_ROOT or $PWD)")
	noAdversarialFlag := flag.Bool("no-adversarial", false, "disable adversarial_review and gate_review tools (use when coexisting with esquisse-mcp)")
	defaultModelFlag := flag.String("default-model", "", "default model for dispatch tools (default: $BOF_DEFAULT_MODEL)")
	flag.Parse()

	// 2. Resolve projectRoot (flag > env > PWD).
	projectRoot := *projectRootFlag
	if projectRoot == "" {
		projectRoot = os.Getenv("BOF_PROJECT_ROOT")
	}
	if projectRoot == "" {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("cannot determine working directory: %v", err)
		}
		projectRoot = pwd
	}

	// Resolve no-adversarial (flag > env).
	noAdversarial := *noAdversarialFlag
	if !noAdversarial && os.Getenv("BOF_NO_ADVERSARIAL") != "" {
		noAdversarial = true
	}

	// Resolve default-model (flag > env).
	defaultModel := *defaultModelFlag
	if defaultModel == "" {
		defaultModel = os.Getenv("BOF_DEFAULT_MODEL")
	}

	// 3. Check for crush in PATH.
	if _, err := exec.LookPath("crush"); err != nil {
		log.Printf("WARN: crush binary not found in PATH — dispatch tools will fail until crush is installed")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Create MCP server.
	server := mcp.NewServer(&mcp.Implementation{Name: "bof-mcp", Version: "0.1.0"}, nil)

	// 5. Register tools.
	registerTools(server, projectRoot, defaultModel, noAdversarial)

	// 6. Run MCP server (blocks until context cancelled or stdin closed).
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Printf("bof-mcp server exited: %v", err)
	}
}
