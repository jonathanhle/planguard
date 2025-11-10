package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jonathanhle/planguard/pkg/config"
	"github.com/jonathanhle/planguard/pkg/parser"
	"github.com/jonathanhle/planguard/pkg/reporter"
	"github.com/jonathanhle/planguard/pkg/scanner"
)

// Version is set at build time
var version = "dev"

func main() {
	// Command-line flags
	configPath := flag.String("config", "", "Path to config file (default: ./.planguard/config.hcl or ~/.planguard/config.hcl)")
	directory := flag.String("directory", ".", "Directory to scan")
	format := flag.String("format", "text", "Output format (text, json, sarif)")
	failOn := flag.String("fail-on", "error", "Fail on severity level (error, warning, info)")
	rulesDir := flag.String("rules-dir", "", "Directory containing rules (default: ~/.planguard/rules)")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Parse()

	if *showVersion {
		fmt.Printf("Planguard v%s\n", version)
		os.Exit(0)
	}

	// Run scan
	exitCode := run(*configPath, *directory, *format, *failOn, *rulesDir)
	os.Exit(exitCode)
}

func run(configPath, directory, format, failOn, rulesDir string) int {
	// Load configuration
	cfg, err := loadConfiguration(configPath, rulesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		return 1
	}

	// Parse Terraform files
	p := parser.NewParser()
	files, err := p.ParseDirectory(directory, cfg.Settings.ExcludePaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing Terraform files: %v\n", err)
		return 1
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No Terraform files found in %s\n", directory)
		return 1
	}

	// Extract resources
	resources, err := parser.ExtractResources(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting resources: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Found %d resources in %d files\n", len(resources), len(files))

	// Create scan context
	ctx := parser.NewScanContext(resources)

	// Run scan
	s := scanner.NewScanner(cfg, cfg.Rules, ctx)
	result, err := s.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during scan: %v\n", err)
		return 1
	}

	// Report results
	rep := reporter.NewReporter(result.Violations, result.FilteredViolations)

	var output string
	switch format {
	case "json":
		output, err = rep.FormatJSON()
	case "sarif":
		output, err = rep.FormatSARIF()
	default:
		output = rep.FormatText()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		return 1
	}

	fmt.Println(output)

	// Determine exit code
	if rep.ShouldFail(failOn) {
		return 1
	}

	return 0
}

func expandHomePath(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(path) == 1 {
		return homeDir, nil
	}

	return homeDir + path[1:], nil
}

func findConfigFile() string {
	// Search order: ./.planguard/config.hcl → ~/.planguard/config.hcl
	candidates := []string{
		"./.planguard/config.hcl",
		"~/.planguard/config.hcl",
	}

	for _, path := range candidates {
		expanded, err := expandHomePath(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(expanded); err == nil {
			return expanded
		}
	}

	return ""
}

func getDefaultRulesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return homeDir + "/.planguard/rules", nil
}

func loadConfiguration(configPath, rulesDir string) (*config.Config, error) {
	// Expand home directory in paths
	if configPath != "" {
		expanded, err := expandHomePath(configPath)
		if err != nil {
			return nil, err
		}
		configPath = expanded
	} else {
		// Auto-search for config file
		configPath = findConfigFile()
	}

	// Expand rules directory path
	if rulesDir != "" {
		expanded, err := expandHomePath(rulesDir)
		if err != nil {
			return nil, err
		}
		rulesDir = expanded
	} else {
		// Use default rules directory
		defaultRulesDir, err := getDefaultRulesDir()
		if err != nil {
			return nil, err
		}
		rulesDir = defaultRulesDir
	}

	// Check if rules directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("rules directory not found: %s\n\nPlease create it and add rules:\n  mkdir -p %s\n  cp -r /path/to/rules/* %s/\n\nOr specify a different location:\n  planguard -rules-dir /path/to/rules", rulesDir, rulesDir, rulesDir)
	}

	var cfg *config.Config
	var err error

	// Load config file if it exists
	if configPath != "" {
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
	} else {
		// Create default config
		cfg = &config.Config{
			Settings: &config.Settings{
				FailOnWarning: false,
				ExcludePaths:  []string{"**/.terraform/**", "**/node_modules/**"},
			},
			Rules:      []config.Rule{},
			Exceptions: []config.Exception{},
			Functions:  []config.Function{},
		}
	}

	// Load rules from directory if not present in config
	if len(cfg.Rules) == 0 && rulesDir != "" {
		rules, err := config.LoadDefaultRules(rulesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load rules from %s: %w", rulesDir, err)
		}
		cfg.Rules = rules
	}

	return cfg, nil
}
