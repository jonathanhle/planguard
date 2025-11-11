package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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
	usePresuppliedRules := flag.String("use-presupplied-rules", "", "Enable presupplied rules (true/false, default: true)")
	presuppliedRulesCategories := flag.String("presupplied-rules-categories", "", "Comma-separated list of presupplied rule categories (aws,azure,common,security,tagging)")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Parse()

	if *showVersion {
		fmt.Printf("Planguard v%s\n", version)
		os.Exit(0)
	}

	// Run scan
	exitCode := run(*configPath, *directory, *format, *failOn, *rulesDir, *usePresuppliedRules, *presuppliedRulesCategories)
	os.Exit(exitCode)
}

func run(configPath, directory, format, failOn, rulesDir string, usePresuppliedRules string, presuppliedRulesCategories string) int {
	// Load configuration
	cfg, err := loadConfiguration(configPath, rulesDir, usePresuppliedRules, presuppliedRulesCategories)
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

func loadConfiguration(configPath, rulesDir string, usePresuppliedRulesStr string, presuppliedRulesCategoriesStr string) (*config.Config, error) {
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
		defaultUsePresuppliedRules := true
		cfg = &config.Config{
			Settings: &config.Settings{
				FailOnWarning:              false,
				ExcludePaths:               []string{"**/.terraform/**", "**/node_modules/**"},
				UsePresuppliedRules:        &defaultUsePresuppliedRules,
				PresuppliedRulesCategories: []string{},
			},
			Rules:      []config.Rule{},
			Exceptions: []config.Exception{},
			Functions:  []config.Function{},
		}
	}

	// Override config settings with CLI flags (only if explicitly provided)
	if usePresuppliedRulesStr != "" {
		usePresuppliedRules := strings.ToLower(usePresuppliedRulesStr) == "true"
		cfg.Settings.UsePresuppliedRules = &usePresuppliedRules
	}

	// Parse comma-separated categories from CLI (only if explicitly provided)
	if presuppliedRulesCategoriesStr != "" {
		categories := []string{}
		for _, cat := range strings.Split(presuppliedRulesCategoriesStr, ",") {
			trimmed := strings.TrimSpace(cat)
			if trimmed != "" {
				categories = append(categories, trimmed)
			}
		}
		if len(categories) > 0 {
			cfg.Settings.PresuppliedRulesCategories = categories
		}
	}

	// Check if we should load presupplied rules
	shouldLoadPresuppliedRules := cfg.Settings.UsePresuppliedRules != nil && *cfg.Settings.UsePresuppliedRules

	// Check if rules directory exists (only if we need to load presupplied rules)
	if shouldLoadPresuppliedRules {
		if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("rules directory not found: %s\n\nPlease create it and add rules:\n  mkdir -p %s\n  cp -r /path/to/rules/* %s/\n\nOr specify a different location:\n  planguard -rules-dir /path/to/rules", rulesDir, rulesDir, rulesDir)
		}
	}

	// Load presupplied rules if enabled and not already present in config
	if len(cfg.Rules) == 0 && shouldLoadPresuppliedRules && rulesDir != "" {
		var rules []config.Rule
		if len(cfg.Settings.PresuppliedRulesCategories) > 0 {
			// Load specific categories
			rules, err = config.LoadDefaultRulesWithCategories(rulesDir, cfg.Settings.PresuppliedRulesCategories)
			if err != nil {
				return nil, fmt.Errorf("failed to load presupplied rules from %s: %w", rulesDir, err)
			}
			fmt.Fprintf(os.Stderr, "Loaded presupplied rules for categories: %s\n", strings.Join(cfg.Settings.PresuppliedRulesCategories, ", "))
		} else {
			// Load all presupplied rules
			rules, err = config.LoadDefaultRules(rulesDir)
			if err != nil {
				return nil, fmt.Errorf("failed to load presupplied rules from %s: %w", rulesDir, err)
			}
		}
		cfg.Rules = rules
	} else if !shouldLoadPresuppliedRules {
		fmt.Fprintf(os.Stderr, "Presupplied rules disabled\n")
	}

	return cfg, nil
}
