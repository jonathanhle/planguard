package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Config represents the guardian configuration
type Config struct {
	Settings   *Settings   `hcl:"settings,block"`
	Rules      []Rule      `hcl:"rule,block"`
	Exceptions []Exception `hcl:"exception,block"`
	Functions  []Function  `hcl:"function,block"`
}

// Settings contains global configuration
type Settings struct {
	FailOnWarning             bool     `hcl:"fail_on_warning,optional"`
	ExcludePaths              []string `hcl:"exclude_paths,optional"`
	UsePresuppliedRules       *bool    `hcl:"use_presupplied_rules,optional"`
	PresuppliedRulesCategories []string `hcl:"presupplied_rules_categories,optional"`
}

// Rule represents a security/compliance rule
type Rule struct {
	ID           string      `hcl:"id,label"`
	Name         string      `hcl:"name"`
	Severity     string      `hcl:"severity"`
	ResourceType string      `hcl:"resource_type"`
	When         *WhenBlock  `hcl:"when,block"`
	Conditions   []Condition `hcl:"condition,block"`
	Message      string      `hcl:"message"`
	Remediation  *string     `hcl:"remediation,optional"`
	References   []string    `hcl:"references,optional"`
}

// WhenBlock represents a conditional execution block
type WhenBlock struct {
	Expression string `hcl:"expression"`
}

// Condition represents a rule condition
type Condition struct {
	Expression string `hcl:"expression"`
}

// Exception represents a rule exception
type Exception struct {
	Rules         []string `hcl:"rules"`
	Paths         []string `hcl:"paths,optional"`
	ResourceNames []string `hcl:"resource_names,optional"`
	Reason        string   `hcl:"reason"`
	ExpiresAt     *string  `hcl:"expires_at,optional"`
	ApprovedBy    string   `hcl:"approved_by"`
	Ticket        *string  `hcl:"ticket,optional"`
}

// Function represents a user-defined function
type Function struct {
	Name       string   `hcl:"name,label"`
	Params     []string `hcl:"params"`
	Expression string   `hcl:"expression"`
}

// Violation represents a rule violation
type Violation struct {
	RuleID       string
	RuleName     string
	Severity     string
	Message      string
	File         string
	Line         int
	Column       int
	ResourceType string
	ResourceName string
	Remediation  string
}

// FilteredViolation represents a violation that was filtered by an exception
type FilteredViolation struct {
	Violation Violation
	Exception Exception
}

// Resource represents a parsed Terraform resource
type Resource struct {
	Type       string
	Name       string
	Attributes map[string]cty.Value
	RawExprs   map[string]hcl.Expression // Raw HCL expressions for function call detection
	File       string
	Line       int
	Column     int
	Labels     []string
}
