# Usage Guide

A guide to configuring and using Gommitlint for Git commit message validation.

## Table of Contents

- [Installation](#installation)
- [Basic Commands](#basic-commands)
- [Configuration](#configuration)
  - [Default Configuration](#default-configuration)
  - [Configuration Discovery](#configuration-discovery)
  - [Custom Configuration](#custom-configuration)
- [Validation Rules](#validation-rules)
- [Output Formats](#output-formats)
- [Integration](#integration)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Installation

### Go Install

```bash
go install github.com/itiquette/gommitlint@latest
```

### Binary Download

```bash
# Linux
curl -L https://github.com/itiquette/gommitlint/releases/latest/download/gommitlint-linux-amd64 -o gommitlint

# macOS  
curl -L https://github.com/itiquette/gommitlint/releases/latest/download/gommitlint-darwin-amd64 -o gommitlint

# Windows
curl -L https://github.com/itiquette/gommitlint/releases/latest/download/gommitlint-windows-amd64.exe -o gommitlint.exe

chmod +x gommitlint
```

### Package Managers

```bash
# Homebrew (macOS/Linux)
brew install itiquette/tap/gommitlint

# Chocolatey (Windows)
choco install gommitlint
```

## Basic Commands

### Validation

```bash
# Validate last commit (default behavior)
gommitlint validate

# Validate specific commit
gommitlint validate --ref=abc123

# Validate commit range
gommitlint validate --range=main..HEAD

# Validate from base branch
gommitlint validate --base-branch=main

# Validate multiple commits from HEAD
gommitlint validate --count=5

# Validate message from file
gommitlint validate --message-file=commit-msg.txt
```

### Git Hooks

```bash
# Install commit-msg hook (recommended)
gommitlint install-hook

# Remove hook
gommitlint remove-hook

# Check hook status
gommitlint status
```

### Help and Information

```bash
# General help
gommitlint --help

# Command-specific help
gommitlint validate --help

# Rule-specific guidance
gommitlint validate --rule-help=conventional
```

## Configuration

### Default Configuration

**Gommitlint works immediately without any configuration.** When no configuration file is found, it uses sensible defaults that follow industry best practices.

#### Rules Enabled by Default

The following rules run automatically with their default settings:

| Rule | Default Behavior | Settings |
|------|------------------|----------|
| `subject` | Validates subject line format | Max 72 characters, lowercase/uppercase validation, no trailing punctuation |
| `conventional` | Enforces Conventional Commits format | Optional scope, standard types (feat, fix, docs, etc.) |
| `signoff` | Requires Signed-off-by line | Must include valid DCO sign-off |
| `signature` | Validates cryptographic signatures | Accepts GPG or SSH signatures |
| `identity` | Verifies committer identity | Checks author and committer match |
| `branchahead` | Limits commits ahead of main | Maximum 50 commits ahead of reference branch |

#### Rules Disabled by Default

These rules require explicit opt-in via configuration:

| Rule | Reason for Default Disabled | Enable With |
|------|----------------------------|-------------|
| `commitbody` | Not all projects require detailed bodies | `rules.enabled: [commitbody]` |
| `jirareference` | Organization-specific requirement | `rules.enabled: [jirareference]` |
| `spell` | Requires dictionary setup | `rules.enabled: [spell]` |

#### Default Settings Summary

```yaml
# This is the effective default configuration
# (you don't need to create this file)
gommitlint:
  message:
    subject:
      max_length: 72
      case: "lower"        # lower|upper
      forbid_endings: ["."]
    body:
      required: false
      min_length: 10
      allow_signoff_only: true
  
  conventional:
    require_scope: false
    types: [feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert]
  
  repo:
    reference_branch: "main"
    max_commits_ahead: 50
  
  signing:
    require_signature: false
    require_verification: false
  
  rules:
    enabled: []     # Empty - using defaults
    disabled: []    # Empty - using defaults
```

#### Zero Configuration Example

```bash
# Just works - no setup required
cd your-git-repo
gommitlint validate

# Output shows which rules run by default
✓ Subject         Subject is 45 characters (max: 72)
✓ Conventional    Valid conventional commit format  
✓ Signoff         Signed-off-by present
✓ Signature       Valid GPG signature
✓ Identity        Author matches committer
✓ BranchAhead     2 commits ahead (max: 50)

PASS: All 6 rules passed
```

### Configuration Discovery

When you want to customize behavior, Gommitlint searches for configuration in priority order:

1. `--config <path>` (explicit file)
2. `.gommitlint.yaml` (current directory)
3. `.gommitlint.yml` (current directory)
4. `$XDG_CONFIG_HOME/gommitlint/config.yaml`
5. `$XDG_CONFIG_HOME/gommitlint/config.yml`
6. Built-in defaults (zero configuration)

### Configuration Options

```bash
# Use specific config file
gommitlint --config /path/to/config.yaml validate

# Ignore all config files (use defaults only)
gommitlint --ignore-config validate

# Show effective configuration
gommitlint config show
```

### Custom Configuration

Create `.gommitlint.yaml` in your repository root to override defaults:

```yaml
gommitlint:
  # Message format requirements
  message:
    subject:
      max_length: 72
      case: "lower"              # lower|upper
      forbid_endings: ["."]
    body:
      required: false            # Require commit body
      min_length: 10             # Minimum body length when required
      allow_signoff_only: true   # Accept DCO-only bodies

  # Rule activation
  rules:
    enabled:
      - commitbody               # Enable body validation
      - jirareference           # Enable JIRA validation
    disabled:
      - signature               # Disable signature validation

  # Conventional commit settings
  conventional:
    require_scope: false
    types: [feat, fix, docs, style, refactor, perf, test, build, ci, chore]

  # JIRA integration
  jira:
    project_prefixes: ["PROJ", "TEAM"]

  # Repository validation
  repo:
    reference_branch: main
    max_commits_ahead: 10

  # Cryptographic signatures
  signing:
    require_signature: false
    require_verification: false
    key_directory: "/path/to/keys"
    allowed_signers: ["user@example.com"]
```

### Rule Priority System

Rules follow explicit priority order:

1. **Explicitly enabled** → Always run (highest priority)
2. **Explicitly disabled** → Never run
3. **Default disabled** → Skip unless enabled
4. **Default enabled** → Run unless disabled

```yaml
rules:
  enabled:
    - commitbody      # Override default-disabled (now runs)
    - spell           # Override default-disabled (now runs)
  disabled:
    - conventional    # Override default-enabled (now skipped)
    - signoff         # Override default-enabled (now skipped)
```

## Validation Rules

### Active Rules Reference

| Rule | Default | Description | Configuration |
|------|---------|-------------|---------------|
| `subject` | ✓ | Subject line length and format | `message.subject.*` |
| `conventional` | ✓ | Conventional Commits format | `conventional.*` |
| `signoff` | ✓ | Signed-off-by requirement | None |
| `signature` | ✓ | GPG/SSH signature validation | `signing.*` |
| `identity` | ✓ | Committer identity validation | None |
| `branchahead` | ✓ | Commits ahead count limit | `repo.max_commits_ahead` |
| `commitbody` | ✗ | Commit body requirements | `message.body.*` |
| `jirareference` | ✗ | JIRA ticket reference requirement | `jira.*` |
| `spell` | ✗ | Spell checking | Requires dictionary setup |

### Rule-Specific Help

```bash
# Get detailed help for any rule
gommitlint validate --rule-help=conventional
gommitlint validate --rule-help=subject
gommitlint validate --rule-help=jirareference
```

## Output Formats

### Progressive Verbosity

```bash
# Minimal output (default)
gommitlint validate

# Technical details for developers
gommitlint validate -v

# Educational guidance for learning
gommitlint validate -vv
```

#### Default Output

```text
✓ Subject         Subject is 42 characters (max: 72)
✗ Conventional    Missing type prefix

FAIL: 1 of 2 rules passed
```

#### Verbose Output (`-v`)

```text
✗ Conventional:
    Error Code: missing_type
    Error Message: Missing type prefix
    Expected: type(scope): description
    Found: "add new feature"
    
    ··················································································
```

#### Extra Verbose Output (`-vv`)

```text
✗ Conventional:
    Error Code: missing_type
    Error Message: Missing type prefix
    Expected: type(scope): description
    Found: "add new feature"

    Help:
    Conventional Commits use a structured format for better automation:
    
    ✓ feat: add user authentication
    ✓ fix(api): resolve memory leak
    ✓ docs: update README examples
    
    ··················································································
```

### Machine-Readable Formats

```bash
# JSON output for automation
gommitlint validate --format=json

# GitHub Actions annotations
gommitlint validate --format=github

# GitLab CI annotations
gommitlint validate --format=gitlab
```

#### JSON Example

```json
{
  "summary": {
    "total_commits": 1,
    "passed_commits": 0,
    "all_passed": false
  },
  "commits": [{
    "commit": {
      "hash": "abc123",
      "subject": "add new feature"
    },
    "rule_results": [{
      "name": "conventional",
      "status": "failed",
      "errors": [{
        "code": "missing_type",
        "message": "Missing type prefix"
      }]
    }],
    "passed": false
  }]
}
```

### Color Control

```bash
# Auto-detect terminal capabilities (default)
gommitlint validate --color=auto

# Force colored output
gommitlint validate --color=always

# Disable colored output
gommitlint validate --color=never

# Environment variable override
NO_COLOR=1 gommitlint validate
```

## Integration

### GitHub Actions

```yaml
name: Validate Commits
on: [pull_request]

jobs:
  validate-commits:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Install Gommitlint
      run: go install github.com/itiquette/gommitlint@latest

    - name: Validate Commits
      run: gommitlint validate --base-branch=origin/${{ github.base_ref }} --format=github
```

### GitLab CI

```yaml
validate-commits:
  stage: quality
  image: golang:1.21-alpine
  before_script:
    - go install github.com/itiquette/gommitlint@latest
  script:
    - gommitlint validate --base-branch=origin/main --format=gitlab
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Validate Commits') {
            steps {
                sh 'go install github.com/itiquette/gommitlint@latest'
                sh 'gommitlint validate --base-branch=origin/main --format=json'
            }
        }
    }
}
```

### Pre-commit Framework

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: gommitlint
        name: Validate commit message
        entry: gommitlint validate --message-file
        language: system
        stages: [commit-msg]
        pass_filenames: false
```

### Manual Git Hook

```bash
# .git/hooks/commit-msg
#!/bin/sh
gommitlint validate --message-file="$1"
```

## Advanced Usage

### Configuration Profiles

```yaml
# .gommitlint.yaml - Production profile
gommitlint:
  message:
    body:
      required: true
      min_length: 30
  rules:
    enabled:
      - commitbody
      - jirareference
      - signature
  signing:
    require_signature: true
    require_verification: true
```

```yaml
# .gommitlint-dev.yaml - Development profile  
gommitlint:
  rules:
    disabled:
      - signature
      - jirareference
```

```bash
# Use development profile
gommitlint --config .gommitlint-dev.yaml validate
```

### Scripting and Automation

```bash
# Check if commits pass validation
if gommitlint validate --format=json | jq -e '.summary.all_passed'; then
    echo "All commits valid"
else
    echo "Validation failures found"
    exit 1
fi

# Get validation summary
gommitlint validate --format=json | jq '.summary'

# Count failed rules
gommitlint validate --format=json | jq '.commits[].rule_results[] | select(.status == "failed") | .name'
```

### Custom Workflows

```bash
# Validate only specific commit types
gommitlint validate --range=main..HEAD | grep -E "(feat|fix)"

# Validate commits with custom message source
echo "feat: add new API endpoint" | gommitlint validate --message-file=-

# Validate commits in specific branch
git checkout feature-branch
gommitlint validate --base-branch=main
```

## Troubleshooting

### Common Issues

#### Rule Not Running

```bash
# Check which rules are active
gommitlint validate -v

# Check effective configuration
gommitlint config show

# Enable specific rule
echo "rules:\n  enabled:\n    - <rule_name>" >> .gommitlint.yaml
```

#### Configuration Not Found

```bash
# Verify configuration search path
gommitlint config show

# Use explicit configuration
gommitlint --config /path/to/config.yaml validate

# Test with defaults only
gommitlint --ignore-config validate
```

#### Git Hook Issues

```bash
# Check hook installation
ls -la .git/hooks/commit-msg

# Reinstall hook
gommitlint remove-hook
gommitlint install-hook

# Test hook manually
echo "test commit" > /tmp/msg
.git/hooks/commit-msg /tmp/msg
```

### Debug Mode

```bash
# Show detailed execution information
gommitlint validate -vv

# Show rule help and configuration
gommitlint validate --rule-help=<rule_name>

# Show configuration resolution
gommitlint config show --verbose
```

### Performance Issues

```bash
# Limit commit range for large repositories
gommitlint validate --count=10

# Use specific commit instead of range
gommitlint validate --ref=HEAD

# Skip expensive rules in CI
gommitlint --config .gommitlint-ci.yaml validate
```

## Configuration Examples

### Minimal Setup

```yaml
gommitlint:
  message:
    subject:
      max_length: 50
```

### Development

```yaml
gommitlint:
  message:
    subject:
      max_length: 72
      case: lower
    body:
      required: true
      min_length: 20
  rules:
    enabled:
      - commitbody
    disabled:
      - signature
      - jirareference
```

### Team Environment

```yaml
gommitlint:
  message:
    body:
      required: true
      min_length: 30
      allow_signoff_only: true
  rules:
    enabled:
      - commitbody
      - jirareference
      - signature
  jira:
    project_prefixes: ["CORP", "TEAM", "PROJ"]
  signing:
    require_signature: true
    require_verification: true
    allowed_signers: ["@company.com"]
```

## Exit Codes

Gommitlint uses standard exit codes for automation:

- `0` • All validations passed successfully
- `1` • Configuration error, invalid arguments, or system failure
- `2` • One or more validation rules failed

### Usage in Scripts

```bash
# Simple success/failure check
if gommitlint validate; then
    echo "Commits are valid"
else
    echo "Validation failed"
fi

# Detailed exit code handling
gommitlint validate
case $? in
    0) echo "All validations passed" ;;
    1) echo "Configuration or system error" ;;
    2) echo "Validation failures found" ;;
esac
```

---

For implementation details and architecture information, see [Architecture Documentation](ARCHITECTURE.md).