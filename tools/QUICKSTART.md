# Quick Start: Converting Terrascan Policies

Get up and running in 5 minutes!

## 1. Setup (One Time)

```bash
# Install Python dependencies
pip install -r tools/requirements.txt

# Set your Anthropic API key
export ANTHROPIC_API_KEY='sk-ant-your-key-here'

# Get a key from: https://console.anthropic.com/
```

## 2. Test with Example

```bash
# Try the example policy (no Terrascan needed)
python tools/convert-terrascan.py tools/example-s3-versioning.rego --dry-run
```

You should see a converted HCL rule!

## 3. Convert Real Terrascan Policies

### Option A: Clone Terrascan

```bash
# Clone the full Terrascan repository
git clone https://github.com/tenable/terrascan.git /tmp/terrascan

# Convert one policy to test
python tools/convert-terrascan.py \
  /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego

# Check the output
cat rules/aws/s3_versioning.hcl
```

### Option B: Download Individual Policies

```bash
# Download a single policy from GitHub
curl -o /tmp/s3Versioning.rego \
  https://raw.githubusercontent.com/tenable/terrascan/master/pkg/policies/opa/rego/aws/aws_s3_bucket/s3Versioning.rego

# Convert it
python tools/convert-terrascan.py /tmp/s3Versioning.rego
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
python tools/convert-terrascan.py \
  /tmp/terrascan/pkg/policies/opa/rego/aws/aws_iam_policy/iamPolicyDocWildcardResource.rego

# RDS encryption
python tools/convert-terrascan.py \
  /tmp/terrascan/pkg/policies/opa/rego/aws/aws_db_instance/rdsEncryptionEnabled.rego

# S3 public access
python tools/convert-terrascan.py \
  /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket/s3EnforceUserACL.rego
```

### Convert by Resource Type

```bash
# All S3 policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_s3_bucket -name "*.rego" | \
  xargs -I {} python tools/convert-terrascan.py {}

# All RDS/Database policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_db_instance -name "*.rego" | \
  xargs -I {} python tools/convert-terrascan.py {}

# All IAM policies
find /tmp/terrascan/pkg/policies/opa/rego/aws/aws_iam_policy -name "*.rego" | \
  xargs -I {} python tools/convert-terrascan.py {}
```

### Custom Output Location

```bash
# Organize by compliance framework
python tools/convert-terrascan.py policy.rego -o rules/compliance/pci-dss/rule1.hcl
python tools/convert-terrascan.py policy.rego -o rules/compliance/hipaa/rule2.hcl

# Organize by severity
python tools/convert-terrascan.py critical.rego -o rules/critical/policy.hcl
python tools/convert-terrascan.py warning.rego -o rules/warnings/policy.hcl
```

## Troubleshooting

### "Module not found: anthropic"

```bash
pip install anthropic
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

If you're converting many policies and hit rate limits:

```bash
# Add delays between conversions
for f in *.rego; do
  python tools/convert-terrascan.py "$f"
  sleep 3  # Wait 3 seconds between requests
done
```

### Conversion Doesn't Look Right

```bash
# Use --dry-run to preview first
python tools/convert-terrascan.py complex-policy.rego --dry-run

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
