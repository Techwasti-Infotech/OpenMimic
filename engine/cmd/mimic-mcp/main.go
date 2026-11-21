package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"openmimic/engine/internal/mcp"
)

func main() {
	stdioFlag := flag.Bool("stdio", true, "Run in stdio mode (default: true for LLM MCP clients)")
	flag.Parse()

	if !*stdioFlag {
		fmt.Fprintln(os.Stderr, "Currently stdio transport is the primary MCP transport for OpenMimic.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := mcp.NewServer(os.Stdin, os.Stdout)
	if err := server.Serve(ctx); err != nil && err != context.Canceled {
		log.Fatalf("OpenMimic MCP Server exited with error: %v", err)
	}
}
