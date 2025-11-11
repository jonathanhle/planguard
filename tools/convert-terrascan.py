#!/usr/bin/env python3
"""
Terrascan to Planguard Converter
Converts Terrascan OPA/Rego policies to Planguard HCL rules using AI.

Usage:
    python convert-terrascan.py <rego_file> [--output <output_file>]

Environment Variables:
    ANTHROPIC_API_KEY: Your Anthropic API key for Claude

Examples:
    # Convert a single policy
    python convert-terrascan.py s3Versioning.rego

    # Specify output file
    python convert-terrascan.py s3Versioning.rego --output rules/aws/s3_versioning.hcl

    # Batch convert a directory
    find terrascan/pkg/policies/opa/rego/aws -name "*.rego" | \
        xargs -I {} python convert-terrascan.py {}
"""

import os
import sys
import json
import argparse
from pathlib import Path

try:
    import anthropic
except ImportError:
    print("Error: anthropic package not installed.")
    print("Install with: pip install anthropic")
    sys.exit(1)


CONVERSION_PROMPT = """You are an expert at converting Terrascan OPA/Rego policies to Planguard HCL rules.

# Planguard HCL Rule Format

Planguard rules use HashiCorp Configuration Language (HCL) with Terraform expression syntax:

```hcl
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
```

## Key Planguard Features:

1. **Direct Resource Access**: Use `self` to access current resource attributes
   - `self.acl`, `self.versioning.enabled`, etc.

2. **Cross-Resource Queries**: Use `resources()` function
   ```hcl
   resources("aws_flow_log")  # Get all flow logs
   resources("aws_*")          # Wildcard matching
   ```

3. **Safe Attribute Access**: Use `has()` and `try()`
   ```hcl
   has(self, "versioning")
   try(self.versioning.enabled, false)
   ```

4. **All Terraform Functions**: length, contains, jsondecode, lookup, etc.

5. **Heredoc for Complex Expressions**: Use `<<-EXPR ... EXPR` for multi-line

## Conversion Guidelines:

1. **Simplify Logic**: Focus on the most common case, not every edge case
2. **Remove Variable Cleanup**: Planguard evaluates after variable resolution
3. **Single Expression**: Combine multiple Rego clauses into one HCL expression with OR logic
4. **Clear Messages**: Simple, actionable violation messages
5. **Resource Type**: Extract from `input.aws_*` pattern

## Common Patterns:

**Rego → HCL Mappings:**
- `input.aws_s3_bucket[_]` → `resource_type = "aws_s3_bucket"`
- `not x` → `!x`
- `bucket.config.versioning` → `self.versioning`
- Pattern matching → `try()` or `has()` functions

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

```rego
{rego_content}
```

# Planguard HCL Rule:
"""


def read_file(file_path):
    """Read file contents."""
    try:
        with open(file_path, 'r') as f:
            return f.read()
    except Exception as e:
        print(f"Error reading file {file_path}: {e}")
        sys.exit(1)


def write_file(file_path, content):
    """Write content to file."""
    try:
        # Create parent directories if needed
        Path(file_path).parent.mkdir(parents=True, exist_ok=True)

        with open(file_path, 'w') as f:
            f.write(content)
        print(f"✅ Converted rule saved to: {file_path}")
    except Exception as e:
        print(f"Error writing file {file_path}: {e}")
        sys.exit(1)


def convert_rego_to_hcl(rego_content, api_key):
    """Convert Rego policy to HCL using Claude API."""
    try:
        client = anthropic.Anthropic(api_key=api_key)

        print("🤖 Converting policy with Claude AI...")

        message = client.messages.create(
            model="claude-sonnet-4-5-20250929",
            max_tokens=4096,
            messages=[
                {
                    "role": "user",
                    "content": CONVERSION_PROMPT.format(rego_content=rego_content)
                }
            ]
        )

        # Extract the HCL content
        hcl_content = message.content[0].text.strip()

        # Clean up any markdown code fences if present
        if hcl_content.startswith("```"):
            lines = hcl_content.split('\n')
            # Remove first and last lines (markdown fences)
            hcl_content = '\n'.join(lines[1:-1])

        return hcl_content

    except anthropic.APIError as e:
        print(f"Error calling Claude API: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected error: {e}")
        sys.exit(1)


def generate_output_path(input_path):
    """Generate output path based on input path."""
    input_file = Path(input_path)

    # Extract rule name from filename (e.g., s3Versioning.rego → s3_versioning.hcl)
    rule_name = input_file.stem

    # Convert camelCase to snake_case
    import re
    rule_name = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', rule_name)
    rule_name = re.sub('([a-z0-9])([A-Z])', r'\1_\2', rule_name).lower()

    output_file = f"{rule_name}.hcl"

    # If input is in a terrascan directory structure, try to preserve it
    if "aws" in str(input_file):
        return f"rules/aws/{output_file}"
    elif "gcp" in str(input_file):
        return f"rules/gcp/{output_file}"
    elif "azure" in str(input_file):
        return f"rules/azure/{output_file}"
    else:
        return f"rules/{output_file}"


def main():
    parser = argparse.ArgumentParser(
        description="Convert Terrascan Rego policies to Planguard HCL rules",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s s3Versioning.rego
  %(prog)s s3Versioning.rego --output rules/aws/s3_versioning.hcl
  %(prog)s policies/*.rego

Environment:
  ANTHROPIC_API_KEY    Your Anthropic API key (required)
        """
    )

    parser.add_argument(
        'rego_file',
        help='Path to Terrascan Rego policy file'
    )

    parser.add_argument(
        '-o', '--output',
        help='Output HCL file path (auto-generated if not specified)'
    )

    parser.add_argument(
        '--dry-run',
        action='store_true',
        help='Print converted rule to stdout instead of saving'
    )

    args = parser.parse_args()

    # Check for API key
    api_key = os.environ.get('ANTHROPIC_API_KEY')
    if not api_key:
        print("❌ Error: ANTHROPIC_API_KEY environment variable not set")
        print("\nGet your API key from: https://console.anthropic.com/")
        print("Then set it: export ANTHROPIC_API_KEY='your-key-here'")
        sys.exit(1)

    # Read Rego file
    print(f"📖 Reading Rego policy: {args.rego_file}")
    rego_content = read_file(args.rego_file)

    # Convert
    hcl_content = convert_rego_to_hcl(rego_content, api_key)

    # Output
    if args.dry_run:
        print("\n" + "="*60)
        print("Converted HCL Rule:")
        print("="*60)
        print(hcl_content)
        print("="*60)
    else:
        output_path = args.output or generate_output_path(args.rego_file)
        write_file(output_path, hcl_content)

        # Print preview
        print("\n📋 Preview:")
        print("-" * 60)
        lines = hcl_content.split('\n')
        for line in lines[:20]:  # Show first 20 lines
            print(line)
        if len(lines) > 20:
            print(f"... ({len(lines) - 20} more lines)")
        print("-" * 60)


if __name__ == '__main__':
    main()
