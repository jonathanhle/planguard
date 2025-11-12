# Quick Start: Converting Terrascan Policies

Get up and running in 5 minutes!

## 1. Setup (One Time)

```bash
# Build the converter
cd tools/convert-terrascan
make build
cd ../..

# Set your Anthropic API key
export ANTHROPIC_API_KEY='sk-ant-your-key-here'

# Get a key from: https://console.anthropic.com/
```

## 2. Test with Example

```bash
# Try the example policy (no Terrascan needed)
./tools/convert-terrascan/convert-terrascan -file tools/example-s3-versioning.rego --dry-run
```

You should see a converted HCL rule!

## 3. Convert Real Terrascan Policies

### Option A: Clone Terrascan

```bash
# Clone the full Terrascan repository
git clone https://github.com/tenable/terrascan.git /tmp/terrascan

# Convert one policy to test
./tools/convert-terrascan/convert-terrascan \
  -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego

# Check the output
cat rules/aws/s3_versioning.hcl
```

### Option B: Download Individual Policies

```bash
# Download a single policy from GitHub
curl -o /tmp/s3Versioning.rego \
  https://raw.githubusercontent.com/tenable/terrascan/master/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego

# Convert it
./tools/convert-terrascan/convert-terrascan -file /tmp/s3Versioning.rego
```

## 4. Batch Convert Multiple Policies

```bash
# Convert all S3 bucket policies
./tools/batch-convert.sh /tmp/terrascan aws_s3_bucket

# Or convert all AWS policies (will take a while!)
./tools/batch-convert.sh /tmp/terrascan
```

## 5. Test Your Converted Rules

```bash
# Create test Terraform code
cat > test.tf << 'EOF'
resource "aws_s3_bucket" "bad_bucket" {
  bucket = "my-test-bucket"
  acl    = "public-read"
}
EOF

# Scan with Planguard
planguard -config .planguard/config.hcl -directory .

# Should catch the public-read ACL!
```

## Common Use Cases

### Convert High-Priority Security Rules

```bash
# IAM wildcard policies
./tools/convert-terrascan/convert-terrascan \
  -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_iam_policy/iamPolicyDocWildcardResource.rego

# RDS encryption
./tools/convert-terrascan/convert-terrascan \
  -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_db_instance/rdsEncryptionEnabled.rego

# S3 public access
./tools/convert-terrascan/convert-terrascan \
  -file /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3EnforceUserACL.rego
```

### Convert by Resource Type

```bash
# All S3 policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket -name "*.rego" | \
  while read f; do ./tools/convert-terrascan/convert-terrascan -file "$f"; sleep 2; done

# All RDS/Database policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_db_instance -name "*.rego" | \
  while read f; do ./tools/convert-terrascan/convert-terrascan -file "$f"; sleep 2; done

# All IAM policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_iam_policy -name "*.rego" | \
  while read f; do ./tools/convert-terrascan/convert-terrascan -file "$f"; sleep 2; done
```

### Custom Output Location

```bash
# Organize by compliance framework
./tools/convert-terrascan/convert-terrascan -file policy.rego -output rules/compliance/pci-dss/rule1.hcl
./tools/convert-terrascan/convert-terrascan -file policy.rego -output rules/compliance/hipaa/rule2.hcl

# Organize by severity
./tools/convert-terrascan/convert-terrascan -file critical.rego -output rules/critical/policy.hcl
./tools/convert-terrascan/convert-terrascan -file warning.rego -output rules/warnings/policy.hcl
```

## Troubleshooting

### "convert-terrascan: command not found"

```bash
# Build it first
cd tools/convert-terrascan && make build && cd ../..

# Or install globally
cd tools/convert-terrascan && make install

# Then use it from anywhere:
convert-terrascan -file policy.rego
```

### "ANTHROPIC_API_KEY not set"

```bash
# Set it in your shell
export ANTHROPIC_API_KEY='sk-ant-your-key-here'

# Or add to ~/.bashrc or ~/.zshrc for persistence
echo "export ANTHROPIC_API_KEY='sk-ant-your-key-here'" >> ~/.bashrc
source ~/.bashrc
```

### Rate Limits

If you're converting many policies and hit rate limits, use the batch script which has built-in delays:

```bash
# Uses 2 second delays between conversions
./tools/batch-convert.sh /tmp/terrascan
```

### Conversion Doesn't Look Right

```bash
# Use --dry-run to preview first
./tools/convert-terrascan/convert-terrascan -file complex-policy.rego --dry-run

# Review the output
# If it needs tweaks, manually edit the saved file
# Or try converting a simpler variant
```

## What's Next?

1. **Review converted rules** - Always check AI output
2. **Test thoroughly** - Run against real Terraform code
3. **Adjust as needed** - Simplify or enhance based on your needs
4. **Add exceptions** - Configure path-based exceptions in `.planguard/config.hcl`
5. **Integrate CI/CD** - Add Planguard to your pipeline

## Cost

Converting policies costs approximately:
- **$0.01-0.05 per policy** (using Claude Sonnet 4.5)
- **$2-5 for 100 policies**
- **$10-25 for 500 policies**

Way cheaper and faster than manual conversion! 🚀

## Getting Help

- 📖 Full docs: `tools/README.md`
- 🐛 Issues: https://github.com/jonathanhle/planguard/issues
- 💡 Questions: Open a discussion on GitHub

## Why Go?

This tool is written in Go to:
- ✅ Keep tooling consistent with the Planguard project
- ✅ Compile to a single binary with no dependencies
- ✅ Make it easy to integrate into the main tool later
- ✅ Provide fast, efficient conversions
