package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iota-uz/devhub"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	configPath := flag.String("config", "devhub.yml", "Path to the devhub.yml config file")
	showVersion := flag.Bool("version", false, "Show version information")
	mcpPort := flag.Int("mcp-port", 8765, "Port for MCP server (Model Context Protocol)")
	flag.Parse()

	if *showVersion {
		_, _ = fmt.Fprintf(os.Stdout, "DevHub CLI\nVersion: %s\nCommit: %s\nBuilt: %s\n", version, commit, date)
		os.Exit(0)
	}

	hub, err := devhub.NewDevHubWithMCPPort(*configPath, *mcpPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create DevHub: %v\n", err)
		os.Exit(1)
	}

	// Create a context that cancels on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	if err := hub.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "DevHub exited with error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("DevHub shut down successfully")
}