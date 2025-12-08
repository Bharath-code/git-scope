package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bharath-code/git-scope/internal/config"
	"github.com/Bharath-code/git-scope/internal/scan"
	"github.com/Bharath-code/git-scope/internal/tui"
)

const version = "0.1.1"

func usage() {
	fmt.Fprintf(os.Stderr, `git-scope v%s — A fast TUI to see the status of all git repositories

Usage:
  git-scope [command] [directories...]

Commands:
  (default)   Launch TUI dashboard
  scan        Scan and print repos (JSON)
  init        Create config file interactively
  help        Show this help

Examples:
  git-scope                    # Scan configured dirs or current dir
  git-scope ~/code ~/work      # Scan specific directories
  git-scope scan .             # Scan current directory (JSON)
  git-scope init               # Setup config interactively

Flags:
`, version)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	configPath := flag.String("config", config.DefaultConfigPath(), "Path to config file")
	flag.Parse()

	args := flag.Args()
	cmd := ""
	dirs := []string{}

	// Parse command and directories
	if len(args) >= 1 {
		switch args[0] {
		case "scan", "tui", "help", "init", "-h", "--help":
			cmd = args[0]
			dirs = args[1:]
		default:
			// Assume it's a directory
			cmd = "tui"
			dirs = args
		}
	}

	// Handle help early
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		usage()
		return
	}

	// Handle init command
	if cmd == "init" {
		runInit()
		return
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Override roots if directories provided via CLI
	if len(dirs) > 0 {
		cfg.Roots = expandDirs(dirs)
	} else if !config.ConfigExists(*configPath) {
		// No config file and no CLI dirs - use smart defaults
		cfg.Roots = getSmartDefaults()
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

// expandDirs converts relative paths and ~ to absolute paths
func expandDirs(dirs []string) []string {
	result := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if d == "." {
			if cwd, err := os.Getwd(); err == nil {
				result = append(result, cwd)
			}
		} else if strings.HasPrefix(d, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				result = append(result, filepath.Join(home, d[2:]))
			}
		} else if filepath.IsAbs(d) {
			result = append(result, d)
		} else {
			if abs, err := filepath.Abs(d); err == nil {
				result = append(result, abs)
			}
		}
	}
	return result
}

// getSmartDefaults returns directories that likely contain git repos
func getSmartDefaults() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		cwd, _ := os.Getwd()
		return []string{cwd}
	}

	// Common developer directories to check
	candidates := []string{
		filepath.Join(home, "code"),
		filepath.Join(home, "Code"),
		filepath.Join(home, "projects"),
		filepath.Join(home, "Projects"),
		filepath.Join(home, "dev"),
		filepath.Join(home, "Dev"),
		filepath.Join(home, "work"),
		filepath.Join(home, "Work"),
		filepath.Join(home, "repos"),
		filepath.Join(home, "Repos"),
		filepath.Join(home, "src"),
		filepath.Join(home, "Developer"),
		filepath.Join(home, "Documents", "GitHub"),
		filepath.Join(home, "Desktop", "projects"),
	}

	found := []string{}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			found = append(found, dir)
		}
	}

	// If no common dirs found, use current directory
	if len(found) == 0 {
		cwd, _ := os.Getwd()
		return []string{cwd}
	}

	return found
}

// runInit creates a config file interactively
func runInit() {
	configPath := config.DefaultConfigPath()
	
	fmt.Println("git-scope init — Setup your configuration")
	fmt.Println()
	
	// Check if config already exists
	if config.ConfigExists(configPath) {
		fmt.Printf("Config file already exists at: %s\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return
		}
	}

	reader := bufio.NewReader(os.Stdin)
	
	// Get directories
	fmt.Println("Enter directories to scan for git repos (one per line, empty line to finish):")
	fmt.Println("Examples: ~/code, ~/projects, ~/work")
	fmt.Println()
	
	dirs := []string{}
	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		dirs = append(dirs, line)
	}

	if len(dirs) == 0 {
		// Suggest detected directories
		detected := getSmartDefaults()
		if len(detected) > 0 {
			fmt.Println("\nNo directories entered. Detected these on your system:")
			for _, d := range detected {
				fmt.Printf("  - %s\n", d)
			}
			fmt.Print("\nUse these? [Y/n]: ")
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer == "" || answer == "y" || answer == "yes" {
				dirs = detected
			} else {
				fmt.Println("No directories configured. Run 'git-scope init' again to set up.")
				return
			}
		}
	}

	// Get editor
	fmt.Print("\nEditor command (default: code): ")
	editor, _ := reader.ReadString('\n')
	editor = strings.TrimSpace(editor)
	if editor == "" {
		editor = "code"
	}

	// Create config
	if err := config.CreateConfig(configPath, dirs, editor); err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	fmt.Printf("\n✅ Config created at: %s\n", configPath)
	fmt.Println("\nRun 'git-scope' to launch the dashboard!")
}
