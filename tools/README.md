# Planguard Conversion Tools

Tools for converting security policies from other formats to Planguard HCL rules.

## Terrascan to Planguard Converter

AI-powered tool written in Go that converts Terrascan OPA/Rego policies to Planguard HCL rules.

### Setup

1. **Build the tool:**
   ```bash
   cd tools/convert-terrascan
   make build
   ```

   Or install to `$GOPATH/bin`:
   ```bash
   cd tools/convert-terrascan
   make install
   ```

2. **Set up API key:**
   ```bash
   export ANTHROPIC_API_KEY='your-api-key-here'
   ```

   Get your API key from: https://console.anthropic.com/

### Usage

#### Convert a Single Policy

```bash
# If installed
convert-terrascan -file path/to/policy.rego

# Or if built locally
./tools/convert-terrascan/convert-terrascan -file path/to/policy.rego
```

This will:
- Read the Rego policy
- Use Claude AI to intelligently convert it to HCL
- Auto-generate an output path (e.g., `rules/aws/s3_versioning.hcl`)
- Save the converted rule

#### Specify Output Location

```bash
convert-terrascan -file s3Versioning.rego -output rules/aws/s3_versioning.hcl
```

#### Preview Without Saving

```bash
convert-terrascan -file s3Versioning.rego --dry-run
```

### Batch Conversion

#### Clone Terrascan Repository

```bash
# Clone terrascan to get policies
git clone https://github.com/tenable/terrascan.git /tmp/terrascan
```

#### Convert All AWS Policies

```bash
# Using the batch script (recommended)
./tools/batch-convert.sh /tmp/terrascan

# Or manually with find and xargs
find /tmp/terrascan/pkg/policies/opa/rego/aws -name "*.rego" | \
  xargs -I {} convert-terrascan -file {}

# Or using a loop for more control
for rego_file in /tmp/terrascan/pkg/policies/opa/rego/aws/**/*.rego; do
  echo "Converting: $rego_file"
  convert-terrascan -file "$rego_file"
  sleep 2  # Rate limiting
done
```

#### Convert Specific Resource Types

```bash
# Just S3 policies
./tools/batch-convert.sh /tmp/terrascan aws_s3_bucket

# RDS policies
./tools/batch-convert.sh /tmp/terrascan aws_db_instance

# IAM policies
./tools/batch-convert.sh /tmp/terrascan aws_iam_policy
```

### Workflow Example

```bash
# 1. Clone Terrascan
git clone https://github.com/tenable/terrascan.git /tmp/terrascan

# 2. Build the converter
cd tools/convert-terrascan && make build && cd ../..

# 3. Set API key
export ANTHROPIC_API_KEY='sk-ant-...'

# 4. Convert a test policy first
./tools/convert-terrascan/convert-terrascan \
  -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego \
  --dry-run

# 5. If it looks good, save it
./tools/convert-terrascan/convert-terrascan \
  -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego

# 6. Test the converted rule
planguard -config .planguard/config.hcl -directory ./test-terraform

# 7. If it works, batch convert more
./tools/batch-convert.sh /tmp/terrascan aws_s3_bucket
```

### What Gets Converted

The AI converter handles:

✅ **Simple attribute checks** - Converted directly
✅ **JSON policy parsing** - Using `jsondecode()`
✅ **Cross-resource relationships** - Using `resources()` function
✅ **Complex logic** - Simplified to common cases
✅ **Metadata** - Extracted into name, message, remediation fields

### Review Converted Rules

**Always review converted rules before using in production:**

1. **Check the logic** - Ensure it matches the original intent
2. **Test with sample code** - Run against test Terraform files
3. **Simplify if needed** - Remove unnecessary complexity
4. **Adjust severity** - Match your organization's requirements
5. **Add exceptions** - Configure path-based exceptions as needed

### Tips

**Start with High-Value Policies:**
```bash
# Security-critical policies first
convert-terrascan -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_iam_policy/iamPolicyDocWildcardResource.rego
convert-terrascan -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3EnforceUserACL.rego
convert-terrascan -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_db_instance/rdsEncryptionEnabled.rego
```

**Iterate on Complex Policies:**
```bash
# Convert with --dry-run first
convert-terrascan -file complex_policy.rego --dry-run

# Review output, then save if good
convert-terrascan -file complex_policy.rego

# Test immediately
planguard -config .planguard/config.hcl -directory ./test-data
```

**Organize by Service:**
```bash
# Keep rules organized
convert-terrascan -file s3Policy.rego -output rules/aws/s3/policy_name.hcl
convert-terrascan -file rdsPolicy.rego -output rules/aws/rds/policy_name.hcl
```

### Troubleshooting

**API Key Issues:**
```bash
# Verify key is set
echo $ANTHROPIC_API_KEY

# Set it if missing
export ANTHROPIC_API_KEY='your-key-here'
```

**Binary Not Found:**
```bash
# Build it
cd tools/convert-terrascan && make build

# Or install globally
cd tools/convert-terrascan && make install
```

**Rate Limits:**
If you hit API rate limits during batch conversion, the batch script has built-in delays. You can adjust the `DELAY` variable in `batch-convert.sh`.

**Complex Policies:**
If a conversion doesn't look right:
1. Use `--dry-run` to review first
2. Manually adjust the converted HCL
3. Simplify the logic if needed
4. Consider splitting into multiple rules

### Cost Estimation

Using Claude Sonnet 4.5:
- **Per conversion:** ~$0.01 - $0.05 (depending on policy complexity)
- **100 policies:** ~$2 - $5
- **500 policies:** ~$10 - $25

Much faster and cheaper than manual conversion!

### Command Line Options

```
convert-terrascan [options]

Options:
  -file string
        Path to Terrascan Rego policy file (required)
  -output string
        Output HCL file path (auto-generated if not specified)
  -o string
        Output HCL file path (shorthand)
  --dry-run
        Print converted rule to stdout instead of saving

Environment:
  ANTHROPIC_API_KEY    Your Anthropic API key (required)

Examples:
  convert-terrascan -file s3Versioning.rego
  convert-terrascan -file s3Versioning.rego -output rules/aws/s3.hcl
  convert-terrascan -file s3Versioning.rego --dry-run
```

### Next Steps

After converting policies:

1. **Test thoroughly** with sample Terraform code
2. **Add to your config** in `.planguard/config.hcl`
3. **Configure exceptions** for known false positives
4. **Integrate into CI/CD** for automated scanning
5. **Share with team** and iterate based on feedback

### Contributing

Found a conversion pattern that doesn't work well?

1. Open an issue with the Rego policy
2. Share the problematic conversion
3. Suggest improvements to the conversion prompt

### License

MIT - Same as Planguard
