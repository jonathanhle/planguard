package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// LoadConfig loads the guardian configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	err := hclsimple.DecodeFile(configPath, nil, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set defaults
	if config.Settings == nil {
		defaultUsePresuppliedRules := true
		config.Settings = &Settings{
			FailOnWarning:              false,
			ExcludePaths:               []string{},
			UsePresuppliedRules:        &defaultUsePresuppliedRules,
			PresuppliedRulesCategories: []string{},
		}
	} else {
		// Set default for UsePresuppliedRules if not specified
		if config.Settings.UsePresuppliedRules == nil {
			defaultUsePresuppliedRules := true
			config.Settings.UsePresuppliedRules = &defaultUsePresuppliedRules
		}
		if config.Settings.PresuppliedRulesCategories == nil {
			config.Settings.PresuppliedRulesCategories = []string{}
		}
	}

	return &config, nil
}

// LoadRules loads rules from one or more HCL files
func LoadRules(rulesPaths []string) ([]Rule, error) {
	var allRules []Rule

	for _, path := range rulesPaths {
		// Check if path is a pattern
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %s: %w", path, err)
		}

		if len(matches) == 0 {
			// Try as literal path
			matches = []string{path}
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			if info.IsDir() {
				// Load all .hcl files in directory
				files, err := filepath.Glob(filepath.Join(match, "*.hcl"))
				if err != nil {
					continue
				}
				matches = append(matches, files...)
				continue
			}

			// Load rules from file
			var fileConfig struct {
				Rules []Rule `hcl:"rule,block"`
			}

			err = hclsimple.DecodeFile(match, nil, &fileConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to load rules from %s: %w", match, err)
			}

			allRules = append(allRules, fileConfig.Rules...)
		}
	}

	return allRules, nil
}

// LoadDefaultRules loads built-in default rules
func LoadDefaultRules(rulesDir string) ([]Rule, error) {
	return LoadDefaultRulesWithCategories(rulesDir, nil)
}

// LoadDefaultRulesWithCategories loads built-in default rules filtered by categories
// Supported categories:
//   - "aws": AWS provider rules (rules/aws/*.hcl)
//   - "azure": Azure provider rules (rules/azure/*.hcl)
//   - "common": All common rules (rules/common/*.hcl)
//   - "security": Security-specific rules (rules/common/security.hcl)
//   - "tagging": Tagging rules (rules/common/tagging.hcl)
// If categories is nil or empty, all rules are loaded
func LoadDefaultRulesWithCategories(rulesDir string, categories []string) ([]Rule, error) {
	if rulesDir == "" {
		// Use embedded rules or skip
		return []Rule{}, nil
	}

	var patterns []string

	// If no categories specified, load all rules (backward compatible)
	if len(categories) == 0 {
		// Load rules from root directory
		rootPattern := filepath.Join(rulesDir, "*.hcl")
		patterns = append(patterns, rootPattern)

		// Load rules from provider subdirectories
		providers := []string{"aws", "azure", "common"}
		for _, provider := range providers {
			pattern := filepath.Join(rulesDir, provider, "*.hcl")
			patterns = append(patterns, pattern)
		}
	} else {
		// Load specific categories
		categoryMap := make(map[string]bool)
		for _, cat := range categories {
			categoryMap[cat] = true
		}

		// Load rules from root directory if any category is specified
		if len(categoryMap) > 0 {
			rootPattern := filepath.Join(rulesDir, "*.hcl")
			patterns = append(patterns, rootPattern)
		}

		// Map categories to file patterns
		if categoryMap["aws"] {
			pattern := filepath.Join(rulesDir, "aws", "*.hcl")
			patterns = append(patterns, pattern)
		}

		if categoryMap["azure"] {
			pattern := filepath.Join(rulesDir, "azure", "*.hcl")
			patterns = append(patterns, pattern)
		}

		if categoryMap["common"] {
			// Load all common rules
			pattern := filepath.Join(rulesDir, "common", "*.hcl")
			patterns = append(patterns, pattern)
		} else {
			// Load specific common rule files
			if categoryMap["security"] {
				pattern := filepath.Join(rulesDir, "common", "security.hcl")
				patterns = append(patterns, pattern)
			}
			if categoryMap["tagging"] {
				pattern := filepath.Join(rulesDir, "common", "tagging.hcl")
				patterns = append(patterns, pattern)
			}
		}
	}

	return LoadRules(patterns)
}
