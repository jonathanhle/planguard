# Planguard Conversion Tools

Tools for converting security policies from other formats to Planguard HCL rules.

## Terrascan to Planguard Converter

AI-powered tool that converts Terrascan OPA/Rego policies to Planguard HCL rules.

### Setup

1. **Install dependencies:**
   ```bash
   pip install anthropic
   ```

2. **Set up API key:**
   ```bash
   export ANTHROPIC_API_KEY='your-api-key-here'
   ```

   Get your API key from: https://console.anthropic.com/

### Usage

#### Convert a Single Policy

```bash
python tools/convert-terrascan.py path/to/policy.rego
```

This will:
- Read the Rego policy
- Use Claude AI to intelligently convert it to HCL
- Auto-generate an output path (e.g., `rules/aws/s3_versioning.hcl`)
- Save the converted rule

#### Specify Output Location

```bash
python tools/convert-terrascan.py s3Versioning.rego --output rules/aws/s3_versioning.hcl
```

#### Preview Without Saving

```bash
python tools/convert-terrascan.py s3Versioning.rego --dry-run
```

### Batch Conversion

#### Clone Terrascan Repository

```bash
# Clone terrascan to get policies
git clone https://github.com/tenable/terrascan.git /tmp/terrascan
```

#### Convert All AWS Policies

```bash
# Using find and xargs
find /tmp/terrascan/pkg/policies/opa/rego/aws -name "*.rego" | \
  xargs -I {} python tools/convert-terrascan.py {}

# Or using a loop for more control
for rego_file in /tmp/terrascan/pkg/policies/opa/rego/aws/**/*.rego; do
  echo "Converting: $rego_file"
  python tools/convert-terrascan.py "$rego_file"
done
```

#### Convert Specific Resource Types

```bash
# Just S3 policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket -name "*.rego" | \
  xargs -I {} python tools/convert-terrascan.py {}

# RDS policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_db_instance -name "*.rego" | \
  xargs -I {} python tools/convert-terrascan.py {}

# IAM policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_iam_policy -name "*.rego" | \
  xargs -I {} python tools/convert-terrascan.py {}
```

### Workflow Example

```bash
# 1. Clone Terrascan
git clone https://github.com/tenable/terrascan.git /tmp/terrascan

# 2. Set API key
export ANTHROPIC_API_KEY='sk-ant-...'

# 3. Convert a test policy first
python tools/convert-terrascan.py \
  /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego \
  --dry-run

# 4. If it looks good, save it
python tools/convert-terrascan.py \
  /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego

# 5. Test the converted rule
planguard -config .planguard/config.hcl -directory ./test-terraform

# 6. If it works, batch convert more
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket -name "*.rego" | \
  head -5 | xargs -I {} python tools/convert-terrascan.py {}
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
python tools/convert-terrascan.py /tmp/terrascan/pkg/policies/opa/rego/aws/aws_iam_policy/iamPolicyDocWildcardResource.rego
python tools/convert-terrascan.py /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3EnforceUserACL.rego
python tools/convert-terrascan.py /tmp/terrascan/pkg/policies/opa/rego/aws/aws_db_instance/rdsEncryptionEnabled.rego
```

**Iterate on Complex Policies:**
```bash
# Convert with --dry-run first
python tools/convert-terrascan.py complex_policy.rego --dry-run

# Review output, then save if good
python tools/convert-terrascan.py complex_policy.rego

# Test immediately
planguard -config .planguard/config.hcl -directory ./test-data
```

**Organize by Service:**
```bash
# Keep rules organized
python tools/convert-terrascan.py s3Policy.rego -o rules/aws/s3/policy_name.hcl
python tools/convert-terrascan.py rdsPolicy.rego -o rules/aws/rds/policy_name.hcl
```

### Troubleshooting

**API Key Issues:**
```bash
# Verify key is set
echo $ANTHROPIC_API_KEY

# Set it if missing
export ANTHROPIC_API_KEY='your-key-here'
```

**Module Not Found:**
```bash
pip install anthropic
```

**Rate Limits:**
If you hit API rate limits during batch conversion, add a delay:
```bash
for rego_file in /tmp/terrascan/pkg/policies/opa/rego/aws/**/*.rego; do
  python tools/convert-terrascan.py "$rego_file"
  sleep 2  # Add 2 second delay between conversions
done
```

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
