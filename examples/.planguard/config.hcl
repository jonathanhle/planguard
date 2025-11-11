# Planguard Configuration
# This file allows you to configure settings, rules, and exceptions

# ====================================================================
# SETTINGS
# ====================================================================
# Global configuration settings for Planguard

settings {
  # Fail on warnings (default: false)
  # Set to true to exit with non-zero code when warnings are found
  fail_on_warning = false

  # Exclude paths from scanning (supports glob patterns)
  exclude_paths = [
    "**/.terraform/**",
    "**/node_modules/**"
  ]

  # Presupplied rules control (default: true)
  # Set to false to disable all presupplied (built-in) rules
  # and only use custom rules defined in this config or rules directory
  use_presupplied_rules = true

  # Presupplied rules categories (optional)
  # If specified, only load presupplied rules from these categories.
  # If empty or not specified, all presupplied rules are loaded.
  # Available categories:
  #   - "aws": AWS provider rules (S3, IAM, RDS, EC2, etc.)
  #   - "azure": Azure provider rules
  #   - "common": All common rules (security + tagging)
  #   - "security": Security-specific rules (exfiltration prevention)
  #   - "tagging": Resource tagging compliance rules
  #
  # Examples:
  #   presupplied_rules_categories = ["security", "aws"]  # Only security and AWS rules
  #   presupplied_rules_categories = ["security"]         # Only security rules
  #   presupplied_rules_categories = []                   # All rules (default)
  #
  # presupplied_rules_categories = ["security", "aws"]
}

# ====================================================================
# EXCEPTION EXAMPLES
# ====================================================================
# Exceptions allow you to suppress specific violations while maintaining
# an audit trail of why they were allowed. All exceptions require:
# - rules: List of rule IDs to except
# - reason: Business justification for the exception
# - approved_by: Who approved the exception
#
# Optional fields:
# - paths: File path patterns (supports wildcards)
# - resource_names: Resource name patterns (supports wildcards)
# - ticket: Reference ticket/issue number
# - expires_at: Expiration date (YYYY-MM-DD format)

# Example 1: Path-based exception
# Allow dangerous patterns in the dangerous-patterns.tf file
# This file is for testing/demonstration purposes
exception {
  rules = [
    "dangerous_http_data_source",
    "dangerous_dns_data_source",
    "dangerous_external_data_source"
  ]
  paths = ["*/dangerous-patterns.tf", "dangerous-patterns.tf"]
  reason = "This file contains intentional security violations for testing the scanner"
  approved_by = "security-team@example.com"
  ticket = "TEST-001"
  expires_at = "2026-12-31"
}

# Example 2: Resource name-based exception
# Allow nonsensitive() in test/demo resources
exception {
  rules = ["dangerous_nonsensitive_function"]
  resource_names = ["*test*", "*demo*", "*example*"]
  reason = "Test and demo resources are not used in production and can expose non-sensitive test data"
  approved_by = "platform-team@example.com"
  ticket = "TEST-002"
}

# Example 3: Allow public S3 bucket in main.tf
# This is for demonstration of exception handling
exception {
  rules = ["aws_s3_public_read"]
  paths = ["*/main.tf", "main.tf"]
  resource_names = ["public_bucket"]
  reason = "Intentional public bucket for static website hosting demonstration"
  approved_by = "architecture-team@example.com"
  ticket = "DEMO-100"
}

# Additional exception examples (commented out - uncomment to use):

# # Example: RDS encryption exception for development
# exception {
#   rules = ["aws_rds_encryption"]
#   paths = ["*/dev/*.tf", "*/development/*.tf"]
#   reason = "Development environments don't require encryption to reduce costs"
#   approved_by = "devops-lead@example.com"
#   ticket = "COST-3344"
#   expires_at = "2025-12-31"
# }
#
# # Example: Legacy infrastructure exception
# exception {
#   rules = ["aws_security_group_ingress_all"]
#   paths = ["*/legacy/*.tf"]
#   reason = "Legacy infrastructure - migration scheduled for Q2 2026"
#   approved_by = "infra-lead@example.com"
#   ticket = "INFRA-9999"
#   expires_at = "2026-06-30"
# }

# ====================================================================
# MONITORING EXCEPTIONS
# ====================================================================
# Best practices for exception management:
#
# 1. Always include 'reason' and 'approved_by' for audit trail
# 2. Use 'ticket' to link to approval documentation
# 3. Set 'expires_at' for temporary exceptions
# 4. Review exceptions quarterly
# 5. Be specific with paths/resource_names when possible
# 6. Document cleanup plans in the reason field
#
# The scan report will show all filtered violations under "EXCEPTED"
# section with full exception details for transparency.
