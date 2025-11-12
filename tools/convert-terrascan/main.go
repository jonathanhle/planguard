package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const conversionPrompt = `You are an expert at converting Terrascan OPA/Rego policies to Planguard HCL rules.

# Planguard HCL Rule Format

Planguard rules use HashiCorp Configuration Language (HCL) with Terraform expression syntax:

` + "```hcl" + `
rule "rule_id" {
  name     = "Human readable name"
  severity = "error"  # or "warning", "info"

  resource_type = "aws_s3_bucket"  # or "*" for all resources

  # Optional: only apply when this condition is true
  when {
    expression = "lookup(self.tags, 'Environment', '') == 'prod'"
  }

  # Main validation logic
  condition {
    expression = "!has(self, 'versioning') || try(self.versioning.enabled, false) != true"
  }

  message = "S3 buckets should enable versioning"

  # Optional remediation guidance
  remediation = <<-EOT
    Enable versioning in your S3 bucket:

    resource "aws_s3_bucket" "example" {
      bucket = "my-bucket"

      versioning {
        enabled = true
      }
    }
  EOT
}
` + "```" + `

## Key Planguard Features:

1. **Direct Resource Access**: Use ` + "`self`" + ` to access current resource attributes
   - ` + "`self.acl`, `self.versioning.enabled`" + `, etc.

2. **Cross-Resource Queries**: Use ` + "`resources()`" + ` function
   ` + "```hcl" + `
   resources("aws_flow_log")  # Get all flow logs
   resources("aws_*")          # Wildcard matching
   ` + "```" + `

3. **Safe Attribute Access**: Use ` + "`has()`" + ` and ` + "`try()`" + `
   ` + "```hcl" + `
   has(self, "versioning")
   try(self.versioning.enabled, false)
   ` + "```" + `

4. **All Terraform Functions**: length, contains, jsondecode, lookup, etc.

5. **Heredoc for Complex Expressions**: Use ` + "`<<-EXPR ... EXPR`" + ` for multi-line

## Conversion Guidelines:

1. **Simplify Logic**: Focus on the most common case, not every edge case
2. **Remove Variable Cleanup**: Planguard evaluates after variable resolution
3. **Single Expression**: Combine multiple Rego clauses into one HCL expression with OR logic
4. **Clear Messages**: Simple, actionable violation messages
5. **Resource Type**: Extract from ` + "`input.aws_*`" + ` pattern

## Common Patterns:

**Rego → HCL Mappings:**
- ` + "`input.aws_s3_bucket[_]`" + ` → ` + "`resource_type = \"aws_s3_bucket\"`" + `
- ` + "`not x`" + ` → ` + "`!x`" + `
- ` + "`bucket.config.versioning`" + ` → ` + "`self.versioning`" + `
- Pattern matching → ` + "`try()`" + ` or ` + "`has()`" + ` functions

# Your Task:

Convert the following Terrascan Rego policy to a Planguard HCL rule.

1. Extract the key check being performed
2. Identify the resource type
3. Simplify the logic to cover common cases
4. Write idiomatic HCL with clear expressions
5. Include helpful message and remediation

Output ONLY the HCL rule, no explanation or markdown formatting.

---

# Terrascan Rego Policy:

` + "```rego" + `
%s
` + "```" + `

# Planguard HCL Rule:`

type Config struct {
	RegoFile  string
	OutputFile string
	DryRun    bool
	APIKey    string
}

func main() {
	cfg := parseFlags()

	// Validate API key
	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: ANTHROPIC_API_KEY environment variable not set")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Get your API key from: https://console.anthropic.com/")
		fmt.Fprintln(os.Stderr, "Then set it: export ANTHROPIC_API_KEY='your-key-here'")
		os.Exit(1)
	}

	// Read Rego file
	fmt.Printf("📖 Reading Rego policy: %s\n", cfg.RegoFile)
	regoContent, err := os.ReadFile(cfg.RegoFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error reading file %s: %v\n", cfg.RegoFile, err)
		os.Exit(1)
	}

	// Convert
	hclContent, err := convertRegoToHCL(string(regoContent), cfg.APIKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error converting policy: %v\n", err)
		os.Exit(1)
	}

	// Output
	if cfg.DryRun {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("Converted HCL Rule:")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println(hclContent)
		fmt.Println(strings.Repeat("=", 60))
	} else {
		// Determine output path
		outputPath := cfg.OutputFile
		if outputPath == "" {
			outputPath = generateOutputPath(cfg.RegoFile)
		}

		// Write file
		if err := writeFile(outputPath, hclContent); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error writing file %s: %v\n", outputPath, err)
			os.Exit(1)
		}

		fmt.Printf("✅ Converted rule saved to: %s\n", outputPath)

		// Print preview
		fmt.Println("\n📋 Preview:")
		fmt.Println(strings.Repeat("-", 60))
		lines := strings.Split(hclContent, "\n")
		for i, line := range lines {
			if i >= 20 {
				fmt.Printf("... (%d more lines)\n", len(lines)-20)
				break
			}
			fmt.Println(line)
		}
		fmt.Println(strings.Repeat("-", 60))
	}
}

func parseFlags() Config {
	cfg := Config{}

	flag.StringVar(&cfg.RegoFile, "file", "", "Path to Terrascan Rego policy file (required)")
	flag.StringVar(&cfg.OutputFile, "output", "", "Output HCL file path (auto-generated if not specified)")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output HCL file path (shorthand)")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Print converted rule to stdout instead of saving")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: convert-terrascan [options]\n\n")
		fmt.Fprintf(os.Stderr, "Convert Terrascan Rego policies to Planguard HCL rules\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  convert-terrascan -file s3Versioning.rego\n")
		fmt.Fprintf(os.Stderr, "  convert-terrascan -file s3Versioning.rego -output rules/aws/s3.hcl\n")
		fmt.Fprintf(os.Stderr, "  convert-terrascan -file s3Versioning.rego --dry-run\n")
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  ANTHROPIC_API_KEY    Your Anthropic API key (required)\n")
	}

	flag.Parse()

	if cfg.RegoFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")

	return cfg
}

func convertRegoToHCL(regoContent, apiKey string) (string, error) {
	fmt.Println("🤖 Converting policy with Claude AI...")

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	prompt := fmt.Sprintf(conversionPrompt, regoContent)

	message, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.F("claude-sonnet-4-20250514"),
		MaxTokens: anthropic.F(int64(4096)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return "", fmt.Errorf("API error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	// Extract text from response
	hclContent := message.Content[0].Text
	hclContent = strings.TrimSpace(hclContent)

	// Clean up any markdown code fences if present
	if strings.HasPrefix(hclContent, "```") {
		lines := strings.Split(hclContent, "\n")
		if len(lines) > 2 {
			// Remove first and last lines (markdown fences)
			hclContent = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	return hclContent, nil
}

func generateOutputPath(inputPath string) string {
	// Extract filename without extension
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	ruleName := strings.TrimSuffix(base, ext)

	// Convert camelCase to snake_case
	ruleName = camelToSnake(ruleName)

	outputFile := fmt.Sprintf("%s.hcl", ruleName)

	// Try to preserve directory structure if it's from terrascan
	if strings.Contains(inputPath, "aws") {
		return filepath.Join("rules", "aws", outputFile)
	} else if strings.Contains(inputPath, "gcp") {
		return filepath.Join("rules", "gcp", outputFile)
	} else if strings.Contains(inputPath, "azure") {
		return filepath.Join("rules", "azure", outputFile)
	}

	return filepath.Join("rules", outputFile)
}

func camelToSnake(s string) string {
	// Insert underscore before uppercase letters
	re1 := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s = re1.ReplaceAllString(s, "${1}_${2}")

	re2 := regexp.MustCompile("([a-z0-9])([A-Z])")
	s = re2.ReplaceAllString(s, "${1}_${2}")

	return strings.ToLower(s)
}

func writeFile(path, content string) error {
	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
