package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.hcl")

	configContent := `
settings {
  fail_on_warning = true
  exclude_paths = ["**/.terraform/**"]
}

rule "test_rule" {
  name     = "Test Rule"
  severity = "error"
  resource_type = "aws_instance"

  condition {
    expression = "true"
  }

  message = "Test message"
}

exception {
  rules = ["test_rule"]
  reason = "Testing"
  approved_by = "test@example.com"
}
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test loading config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify settings
	if cfg.Settings == nil {
		t.Error("Settings should not be nil")
	}
	if !cfg.Settings.FailOnWarning {
		t.Error("FailOnWarning should be true")
	}
	if len(cfg.Settings.ExcludePaths) != 1 {
		t.Errorf("Expected 1 exclude path, got %d", len(cfg.Settings.ExcludePaths))
	}

	// Verify rules
	if len(cfg.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].ID != "test_rule" {
		t.Errorf("Rule ID = %s, want test_rule", cfg.Rules[0].ID)
	}

	// Verify exceptions
	if len(cfg.Exceptions) != 1 {
		t.Errorf("Expected 1 exception, got %d", len(cfg.Exceptions))
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create temporary empty config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.hcl")

	err := os.WriteFile(configPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test loading config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify defaults are set
	if cfg.Settings == nil {
		t.Fatal("Settings should not be nil")
	}
	if cfg.Settings.FailOnWarning != false {
		t.Error("Default FailOnWarning should be false")
	}
	if cfg.Settings.ExcludePaths == nil {
		t.Error("Default ExcludePaths should not be nil")
	}
}

func TestLoadConfigInvalidPath(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.hcl")
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
}

func TestLoadConfigInvalidHCL(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.hcl")

	invalidContent := `
this is not valid HCL {{{
`

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid HCL")
	}
}

func TestLoadRules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test rule file
	ruleFile := filepath.Join(tmpDir, "test.hcl")
	ruleContent := `
rule "test_rule" {
  name     = "Test Rule"
  severity = "error"
  resource_type = "aws_instance"

  condition {
    expression = "true"
  }

  message = "Test message"
}

rule "second_rule" {
  name     = "Second Rule"
  severity = "warning"
  resource_type = "aws_s3_bucket"

  condition {
    expression = "false"
  }

  message = "Second message"
}
`

	err := os.WriteFile(ruleFile, []byte(ruleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create rule file: %v", err)
	}

	// Test loading rules
	rules, err := LoadRules([]string{ruleFile})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}

	if rules[0].ID != "test_rule" {
		t.Errorf("First rule ID = %s, want test_rule", rules[0].ID)
	}

	if rules[1].ID != "second_rule" {
		t.Errorf("Second rule ID = %s, want second_rule", rules[1].ID)
	}
}

func TestLoadRulesFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple rule files
	file1 := filepath.Join(tmpDir, "rules1.hcl")
	file2 := filepath.Join(tmpDir, "rules2.hcl")

	rule1Content := `
rule "rule_1" {
  name     = "Rule 1"
  severity = "error"
  resource_type = "aws_instance"
  condition {
    expression = "true"
  }
  message = "Message 1"
}
`

	rule2Content := `
rule "rule_2" {
  name     = "Rule 2"
  severity = "warning"
  resource_type = "aws_s3_bucket"
  condition {
    expression = "true"
  }
  message = "Message 2"
}
`

	if err := os.WriteFile(file1, []byte(rule1Content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(rule2Content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test loading rules from directory using glob pattern
	pattern := filepath.Join(tmpDir, "*.hcl")
	rules, err := LoadRules([]string{pattern})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}
}

func TestLoadRulesGlobPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with different extensions
	hclFile := filepath.Join(tmpDir, "test.hcl")
	txtFile := filepath.Join(tmpDir, "test.txt")

	ruleContent := `
rule "glob_test" {
  name     = "Glob Test"
  severity = "error"
  resource_type = "aws_instance"
  condition {
    expression = "true"
  }
  message = "Test"
}
`

	if err := os.WriteFile(hclFile, []byte(ruleContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(txtFile, []byte("not hcl"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test loading with glob pattern
	pattern := filepath.Join(tmpDir, "*.hcl")
	rules, err := LoadRules([]string{pattern})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
}

func TestLoadRulesNonexistent(t *testing.T) {
	rules, err := LoadRules([]string{"/nonexistent/path/*.hcl"})
	if err != nil {
		t.Fatalf("LoadRules() should not error on nonexistent path: %v", err)
	}

	// Should return empty slice, not error
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules for nonexistent path, got %d", len(rules))
	}
}

func TestLoadRulesInvalidHCL(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.hcl")

	err := os.WriteFile(invalidFile, []byte("invalid {{{"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	_, err = LoadRules([]string{invalidFile})
	if err == nil {
		t.Error("Expected error for invalid HCL in rules")
	}
}

func TestLoadDefaultRules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create provider subdirectories
	awsDir := filepath.Join(tmpDir, "aws")
	azureDir := filepath.Join(tmpDir, "azure")
	commonDir := filepath.Join(tmpDir, "common")

	if err := os.MkdirAll(awsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(azureDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(commonDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create root rule
	rootRule := filepath.Join(tmpDir, "root.hcl")
	rootContent := `
rule "root_rule" {
  name     = "Root Rule"
  severity = "error"
  resource_type = "*"
  condition {
    expression = "true"
  }
  message = "Root"
}
`
	if err := os.WriteFile(rootRule, []byte(rootContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create AWS rule
	awsRule := filepath.Join(awsDir, "aws.hcl")
	awsContent := `
rule "aws_rule" {
  name     = "AWS Rule"
  severity = "error"
  resource_type = "aws_*"
  condition {
    expression = "true"
  }
  message = "AWS"
}
`
	if err := os.WriteFile(awsRule, []byte(awsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create Azure rule
	azureRule := filepath.Join(azureDir, "azure.hcl")
	azureContent := `
rule "azure_rule" {
  name     = "Azure Rule"
  severity = "error"
  resource_type = "azurerm_*"
  condition {
    expression = "true"
  }
  message = "Azure"
}
`
	if err := os.WriteFile(azureRule, []byte(azureContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create common rule
	commonRule := filepath.Join(commonDir, "common.hcl")
	commonContent := `
rule "common_rule" {
  name     = "Common Rule"
  severity = "warning"
  resource_type = "*"
  condition {
    expression = "true"
  }
  message = "Common"
}
`
	if err := os.WriteFile(commonRule, []byte(commonContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test loading default rules
	rules, err := LoadDefaultRules(tmpDir)
	if err != nil {
		t.Fatalf("LoadDefaultRules() error = %v", err)
	}

	// Should load root + all subdirectories
	if len(rules) != 4 {
		t.Errorf("Expected 4 rules (root + 3 providers), got %d", len(rules))
	}

	// Verify we got rules from all locations
	ruleIDs := make(map[string]bool)
	for _, rule := range rules {
		ruleIDs[rule.ID] = true
	}

	expectedRules := []string{"root_rule", "aws_rule", "azure_rule", "common_rule"}
	for _, expected := range expectedRules {
		if !ruleIDs[expected] {
			t.Errorf("Missing expected rule: %s", expected)
		}
	}
}

func TestLoadDefaultRulesEmpty(t *testing.T) {
	rules, err := LoadDefaultRules("")
	if err != nil {
		t.Fatalf("LoadDefaultRules() error = %v", err)
	}

	if len(rules) != 0 {
		t.Errorf("Expected 0 rules for empty rulesDir, got %d", len(rules))
	}
}

func TestLoadDefaultRulesNonexistent(t *testing.T) {
	rules, err := LoadDefaultRules("/nonexistent/path")
	if err != nil {
		t.Fatalf("LoadDefaultRules() should not error: %v", err)
	}

	if len(rules) != 0 {
		t.Errorf("Expected 0 rules for nonexistent path, got %d", len(rules))
	}
}

func TestLoadRulesMultiplePaths(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Create rule in first directory
	file1 := filepath.Join(tmpDir1, "rule1.hcl")
	content1 := `
rule "multi_1" {
  name     = "Multi 1"
  severity = "error"
  resource_type = "aws_instance"
  condition {
    expression = "true"
  }
  message = "Test 1"
}
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}

	// Create rule in second directory
	file2 := filepath.Join(tmpDir2, "rule2.hcl")
	content2 := `
rule "multi_2" {
  name     = "Multi 2"
  severity = "warning"
  resource_type = "aws_s3_bucket"
  condition {
    expression = "true"
  }
  message = "Test 2"
}
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}

	// Load rules from both paths
	rules, err := LoadRules([]string{file1, file2})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules from multiple paths, got %d", len(rules))
	}
}

func TestLoadRulesEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.hcl")

	// Create empty HCL file
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules([]string{emptyFile})
	if err != nil {
		t.Fatalf("LoadRules() should handle empty file: %v", err)
	}

	if len(rules) != 0 {
		t.Errorf("Expected 0 rules from empty file, got %d", len(rules))
	}
}

func TestLoadRulesWithRuleRemediation(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "test.hcl")

	content := `
rule "with_remediation" {
  name     = "With Remediation"
  severity = "error"
  resource_type = "*"
  condition {
    expression = "true"
  }
  message = "Test"
  remediation = "Fix it like this..."
}

rule "without_remediation" {
  name     = "Without Remediation"
  severity = "warning"
  resource_type = "*"
  condition {
    expression = "false"
  }
  message = "Test 2"
}
`

	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules([]string{ruleFile})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(rules))
	}

	// Check that remediation was loaded
	if rules[0].Remediation == nil {
		t.Error("Expected remediation to be set for first rule")
	}

	if rules[1].Remediation != nil {
		t.Error("Expected remediation to be nil for second rule")
	}
}

func TestLoadRulesWithWhenBlock(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "test.hcl")

	content := `
rule "with_when" {
  name     = "With When"
  severity = "error"
  resource_type = "*"

  when {
    expression = "true"
  }

  condition {
    expression = "false"
  }

  message = "Test"
}
`

	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules([]string{ruleFile})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	if rules[0].When == nil {
		t.Error("Expected when block to be set")
	}
}

func TestLoadRulesWithReferences(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "test.hcl")

	content := `
rule "with_refs" {
  name     = "With References"
  severity = "error"
  resource_type = "*"

  condition {
    expression = "true"
  }

  message = "Test"
  references = ["https://example.com/rule", "https://docs.example.com"]
}
`

	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadRules([]string{ruleFile})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	if len(rules[0].References) != 2 {
		t.Errorf("Expected 2 references, got %d", len(rules[0].References))
	}
}

func TestLoadDefaultRulesWithCategories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create provider subdirectories
	awsDir := filepath.Join(tmpDir, "aws")
	azureDir := filepath.Join(tmpDir, "azure")
	commonDir := filepath.Join(tmpDir, "common")

	if err := os.MkdirAll(awsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(azureDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(commonDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create root rule
	rootRule := filepath.Join(tmpDir, "root.hcl")
	rootContent := `
rule "root_rule" {
  name     = "Root Rule"
  severity = "error"
  resource_type = "*"
  condition {
    expression = "true"
  }
  message = "Root"
}
`
	if err := os.WriteFile(rootRule, []byte(rootContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create AWS rule
	awsRule := filepath.Join(awsDir, "aws.hcl")
	awsContent := `
rule "aws_rule" {
  name     = "AWS Rule"
  severity = "error"
  resource_type = "aws_*"
  condition {
    expression = "true"
  }
  message = "AWS"
}
`
	if err := os.WriteFile(awsRule, []byte(awsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create Azure rule
	azureRule := filepath.Join(azureDir, "azure.hcl")
	azureContent := `
rule "azure_rule" {
  name     = "Azure Rule"
  severity = "error"
  resource_type = "azurerm_*"
  condition {
    expression = "true"
  }
  message = "Azure"
}
`
	if err := os.WriteFile(azureRule, []byte(azureContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create security rule
	securityRule := filepath.Join(commonDir, "security.hcl")
	securityContent := `
rule "security_rule" {
  name     = "Security Rule"
  severity = "error"
  resource_type = "*"
  condition {
    expression = "true"
  }
  message = "Security"
}
`
	if err := os.WriteFile(securityRule, []byte(securityContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create tagging rule
	taggingRule := filepath.Join(commonDir, "tagging.hcl")
	taggingContent := `
rule "tagging_rule" {
  name     = "Tagging Rule"
  severity = "warning"
  resource_type = "*"
  condition {
    expression = "true"
  }
  message = "Tagging"
}
`
	if err := os.WriteFile(taggingRule, []byte(taggingContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		categories []string
		wantRules  []string
	}{
		{
			name:       "All rules (nil categories)",
			categories: nil,
			wantRules:  []string{"root_rule", "aws_rule", "azure_rule", "security_rule", "tagging_rule"},
		},
		{
			name:       "All rules (empty categories)",
			categories: []string{},
			wantRules:  []string{"root_rule", "aws_rule", "azure_rule", "security_rule", "tagging_rule"},
		},
		{
			name:       "AWS only",
			categories: []string{"aws"},
			wantRules:  []string{"root_rule", "aws_rule"},
		},
		{
			name:       "Azure only",
			categories: []string{"azure"},
			wantRules:  []string{"root_rule", "azure_rule"},
		},
		{
			name:       "Security only",
			categories: []string{"security"},
			wantRules:  []string{"root_rule", "security_rule"},
		},
		{
			name:       "Tagging only",
			categories: []string{"tagging"},
			wantRules:  []string{"root_rule", "tagging_rule"},
		},
		{
			name:       "Common (all common rules)",
			categories: []string{"common"},
			wantRules:  []string{"root_rule", "security_rule", "tagging_rule"},
		},
		{
			name:       "AWS and security",
			categories: []string{"aws", "security"},
			wantRules:  []string{"root_rule", "aws_rule", "security_rule"},
		},
		{
			name:       "Multiple categories",
			categories: []string{"aws", "azure", "tagging"},
			wantRules:  []string{"root_rule", "aws_rule", "azure_rule", "tagging_rule"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := LoadDefaultRulesWithCategories(tmpDir, tt.categories)
			if err != nil {
				t.Fatalf("LoadDefaultRulesWithCategories() error = %v", err)
			}

			if len(rules) != len(tt.wantRules) {
				t.Errorf("Expected %d rules, got %d", len(tt.wantRules), len(rules))
			}

			// Verify we got the expected rules
			ruleIDs := make(map[string]bool)
			for _, rule := range rules {
				ruleIDs[rule.ID] = true
			}

			for _, expected := range tt.wantRules {
				if !ruleIDs[expected] {
					t.Errorf("Missing expected rule: %s", expected)
				}
			}
		})
	}
}

func TestLoadConfigWithPresuppliedRulesSettings(t *testing.T) {
	tests := []struct {
		name                   string
		configContent          string
		wantUsePresupplied     bool
		wantCategories         []string
		wantUsePresuppliedNil  bool
	}{
		{
			name: "use_presupplied_rules = false",
			configContent: `
settings {
  use_presupplied_rules = false
}
`,
			wantUsePresupplied: false,
			wantCategories:     []string{},
		},
		{
			name: "use_presupplied_rules = true",
			configContent: `
settings {
  use_presupplied_rules = true
}
`,
			wantUsePresupplied: true,
			wantCategories:     []string{},
		},
		{
			name: "presupplied_rules_categories with single category",
			configContent: `
settings {
  presupplied_rules_categories = ["security"]
}
`,
			wantUsePresupplied: true, // default
			wantCategories:     []string{"security"},
		},
		{
			name: "presupplied_rules_categories with multiple categories",
			configContent: `
settings {
  presupplied_rules_categories = ["security", "aws", "tagging"]
}
`,
			wantUsePresupplied: true, // default
			wantCategories:     []string{"security", "aws", "tagging"},
		},
		{
			name: "Both settings specified",
			configContent: `
settings {
  use_presupplied_rules = true
  presupplied_rules_categories = ["aws"]
}
`,
			wantUsePresupplied: true,
			wantCategories:     []string{"aws"},
		},
		{
			name: "No presupplied rules settings (defaults)",
			configContent: `
settings {
  fail_on_warning = false
}
`,
			wantUsePresupplied: true, // default
			wantCategories:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.hcl")

			err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config: %v", err)
			}

			cfg, err := LoadConfig(configPath)
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			// Verify UsePresuppliedRules
			if cfg.Settings.UsePresuppliedRules == nil {
				t.Error("UsePresuppliedRules should not be nil")
			} else if *cfg.Settings.UsePresuppliedRules != tt.wantUsePresupplied {
				t.Errorf("UsePresuppliedRules = %v, want %v", *cfg.Settings.UsePresuppliedRules, tt.wantUsePresupplied)
			}

			// Verify PresuppliedRulesCategories
			if len(cfg.Settings.PresuppliedRulesCategories) != len(tt.wantCategories) {
				t.Errorf("PresuppliedRulesCategories length = %d, want %d",
					len(cfg.Settings.PresuppliedRulesCategories), len(tt.wantCategories))
			}

			for i, cat := range tt.wantCategories {
				if i >= len(cfg.Settings.PresuppliedRulesCategories) {
					t.Errorf("Missing category: %s", cat)
					continue
				}
				if cfg.Settings.PresuppliedRulesCategories[i] != cat {
					t.Errorf("Category[%d] = %s, want %s",
						i, cfg.Settings.PresuppliedRulesCategories[i], cat)
				}
			}
		})
	}
}

func TestLoadConfigPresuppliedRulesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.hcl")

	// Empty config should get defaults
	err := os.WriteFile(configPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify defaults
	if cfg.Settings == nil {
		t.Fatal("Settings should not be nil")
	}

	if cfg.Settings.UsePresuppliedRules == nil {
		t.Error("UsePresuppliedRules should have default value")
	} else if *cfg.Settings.UsePresuppliedRules != true {
		t.Error("Default UsePresuppliedRules should be true")
	}

	if cfg.Settings.PresuppliedRulesCategories == nil {
		t.Error("PresuppliedRulesCategories should not be nil")
	}
	if len(cfg.Settings.PresuppliedRulesCategories) != 0 {
		t.Errorf("Default PresuppliedRulesCategories should be empty, got %d items",
			len(cfg.Settings.PresuppliedRulesCategories))
	}
}
