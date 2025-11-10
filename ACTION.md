# Planguard GitHub Action

Use Planguard as a GitHub Action in your CI/CD workflows to scan Terraform code for security and compliance issues.

## Quick Start

```yaml
name: Terraform Security Scan
on: [pull_request]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Planguard Scan
        uses: jonathanhle/planguard@v1
        with:
          directory: terraform/
          fail-on: error
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `config` | Path to planguard config file | No | `.planguard/config.hcl` |
| `directory` | Directory to scan | No | `.` |
| `fail-on` | Severity level to fail on (`error`, `warning`, `info`) | No | `error` |
| `format` | Output format (`text`, `json`, `sarif`) | No | `text` |
| `rules-dir` | Directory containing custom rules | No | Built-in rules |

## Outputs

| Output | Description |
|--------|-------------|
| `violations` | JSON array of violations found |
| `passed` | Whether the scan passed (`true`/`false`) |

## Usage Examples

### Example 1: Basic Scan

Scan Terraform code and fail on errors:

```yaml
- name: Scan Terraform
  uses: jonathanhle/planguard@v1
  with:
    directory: terraform/
    fail-on: error
```

### Example 2: Custom Configuration

Use a custom config file with exceptions:

```yaml
- name: Scan with custom config
  uses: jonathanhle/planguard@v1
  with:
    config: .planguard/config.hcl
    directory: terraform/
    fail-on: warning
```

### Example 3: JSON Output

Get JSON output for further processing:

```yaml
- name: Scan and output JSON
  id: planguard
  uses: jonathanhle/planguard@v1
  with:
    directory: terraform/
    format: json

- name: Process violations
  if: always()
  run: |
    echo "Violations: ${{ steps.planguard.outputs.violations }}"
    echo "Passed: ${{ steps.planguard.outputs.passed }}"
```

### Example 4: SARIF for GitHub Security Tab

Upload results to GitHub Security tab:

```yaml
jobs:
  security-scan:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4

      - name: Run Planguard
        uses: jonathanhle/planguard@v1
        with:
          directory: terraform/
          format: sarif
        continue-on-error: true

      - name: Upload to Security Tab
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: planguard-results.sarif
```

### Example 5: Multiple Directories

Scan multiple directories in parallel:

```yaml
jobs:
  scan:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        dir: [terraform/prod, terraform/staging, terraform/dev]
    steps:
      - uses: actions/checkout@v4

      - name: Scan ${{ matrix.dir }}
        uses: jonathanhle/planguard@v1
        with:
          directory: ${{ matrix.dir }}
```

### Example 6: PR Comments

Comment on PRs with scan results:

```yaml
- name: Scan Terraform
  id: scan
  uses: jonathanhle/planguard@v1
  with:
    directory: terraform/
    format: json
  continue-on-error: true

- name: Comment on PR
  if: github.event_name == 'pull_request'
  uses: actions/github-script@v7
  with:
    script: |
      const violations = ${{ steps.scan.outputs.violations }};
      const body = violations.length > 0
        ? `⚠️ Planguard found ${violations.length} violation(s)`
        : '✅ Planguard scan passed';

      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: body
      });
```

### Example 7: Fail Only on Specific Severities

Warn on all violations but only fail on errors:

```yaml
- name: Scan (warning mode)
  uses: jonathanhle/planguard@v1
  with:
    directory: terraform/
    fail-on: error  # Won't fail on warnings
```

### Example 8: Custom Rules

Use custom rules from your repository:

```yaml
- name: Scan with custom rules
  uses: jonathanhle/planguard@v1
  with:
    directory: terraform/
    config: .planguard/config.hcl
    rules-dir: .planguard/rules
```

## Complete Example Workflow

```yaml
name: Terraform Security

on:
  pull_request:
    paths:
      - 'terraform/**'
  push:
    branches: [main]

jobs:
  planguard-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
      pull-requests: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Planguard
        id: planguard
        uses: jonathanhle/planguard@v1
        with:
          config: .planguard/config.hcl
          directory: terraform/
          format: sarif
          fail-on: error
        continue-on-error: true

      - name: Upload SARIF results
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: planguard-results.sarif

      - name: Comment on PR
        if: github.event_name == 'pull_request' && !steps.planguard.outputs.passed
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '⚠️ Planguard found security violations. Check the Security tab for details.'
            });

      - name: Fail if violations found
        if: steps.planguard.outputs.passed == 'false'
        run: exit 1
```

## Tips

1. **Use `continue-on-error: true`** if you want to upload SARIF even when violations are found
2. **Set appropriate permissions** for SARIF uploads (`security-events: write`)
3. **Use matrix strategy** to scan multiple directories in parallel
4. **Cache rules** if using custom rules from external sources
5. **Pin to a specific version** (`@v1.0.0`) for production workflows

## Troubleshooting

### No rules found

If you see "rules directory not found", ensure your config file exists and rules are available:

```yaml
- name: Setup Planguard config
  run: |
    mkdir -p .planguard
    # Copy your config here

- name: Run Planguard
  uses: jonathanhle/planguard@v1
```

### SARIF upload fails

Ensure you have the correct permissions:

```yaml
permissions:
  security-events: write
```

### Action uses too much disk space

The action builds from source. For faster runs, consider using the pre-built Docker image:

```yaml
- name: Run Planguard
  uses: docker://ghcr.io/jonathanhle/planguard:latest
  with:
    args: -directory terraform/ -format text
```

## Support

- Documentation: https://github.com/jonathanhle/planguard
- Issues: https://github.com/jonathanhle/planguard/issues
- Discussions: https://github.com/jonathanhle/planguard/discussions
