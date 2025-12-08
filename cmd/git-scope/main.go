package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Bharath-code/git-scope/internal/config"
	"github.com/Bharath-code/git-scope/internal/scan"
	"github.com/Bharath-code/git-scope/internal/tui"
)

const version = "0.1.0"

func usage() {
	fmt.Fprintf(os.Stderr, `git-scope v%s — A fast TUI to see the status of all git repositories

Usage:
  git-scope [command]

Commands:
  (default)   Launch TUI dashboard
  scan        Scan configured roots and print repos (JSON)
  tui         Launch TUI dashboard (same as default)
  help        Show this help

Flags:
`, version)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	configPath := flag.String("config", config.DefaultConfigPath(), "Path to config file")
	flag.Parse()

	// Default command is "tui" if no args provided
	cmd := "tui"
	if flag.NArg() >= 1 {
		cmd = flag.Arg(0)
	}

	// Handle help early
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		usage()
		return
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	switch cmd {
	case "scan":
		repos, err := scan.ScanRoots(cfg.Roots, cfg.Ignore)
		if err != nil {
			log.Fatalf("scan error: %v", err)
		}
		if err := scan.PrintJSON(os.Stdout, repos); err != nil {
			log.Fatalf("print error: %v", err)
		}

	case "tui", "":
		if err := tui.Run(cfg); err != nil {
			log.Fatalf("tui error: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}
