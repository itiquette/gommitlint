# Gommitlint

[![GitHub Release](https://img.shields.io/github/v/tag/itiquette/gommitlint?style=for-the-badge&color=blue)](https://github.com/itiquette/gommitlint/releases)  

[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/itiquette/gommitlint/badge?style=for-the-badge)](https://scorecard.dev/viewer/?uri=github.com/itiquette/gommitlint)  

[![License](https://img.shields.io/badge/License-EUPL%201.2-blue?style=for-the-badge)](LICENSE)
[![REUSE Compliance](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.reuse.software%2Fstatus%2Fgithub.com%2Fitiquette%2Fgommitlint&query=status&style=for-the-badge&label=REUSE)](https://api.reuse.software/info/github.com/itiquette/gommitlint)

Gommitlint validates commit messages against best practices. Keep a clean Git history.

NOTE: In heavy dev-phase, don't use seriously yet.

## Features

- **Zero Configuration** • Works out-of-the-box with sensible defaults
- **Conventional Commits** • Enforces standard commit message format
- **CI/CD Ready** • Multiple output formats for automation
- **Git Hooks** • Automatic validation before commits
- **Flexible Rules** • Enable/disable validation as needed
- **Progressive Disclosure** • Different verbosity levels for different use cases

## Quick Start

### Installation

```bash
# Install with Go
go install github.com/itiquette/gommitlint@latest

# Or download binary from releases
curl -L https://github.com/itiquette/gommitlint/releases/latest/download/gommitlint-linux-amd64 -o gommitlint
chmod +x gommitlint
```

### Basic Usage

```bash
# Validate last commit
gommitlint validate

# Validate commit range
gommitlint validate --range=main..HEAD

# Install Git hook for automatic validation
gommitlint install-hook
```

### Example Output

```text
✓ Subject         Subject is 42 characters (max: 72)
✓ Conventional    Valid conventional commit format
✗ Signoff         Missing Signed-off-by line

FAIL: 1 of 3 rules passed
```

## Common Use Cases

### Git Hook Integration

```bash
# Automatic validation before each commit
gommitlint install-hook
```

### CI/CD Pipeline

```yaml
# GitHub Actions
- name: Validate commits
  run: |
    go install github.com/itiquette/gommitlint@latest
    gommitlint validate --base-branch=origin/${{ github.base_ref }} --format=github
```

### Team Configuration

Create `.gommitlint.yaml` in your repository:

```yaml
gommitlint:
  message:
    subject:
      max_length: 50
  rules:
    enabled:
      - commitbody      # Require detailed commit messages
    disabled:
      - signature       # Don't require GPG signatures
```

## Validation Rules

| Rule | Description | Default |
|------|-------------|---------|
| `subject` | Subject line length and format | ✓ |
| `conventional` | [Conventional Commits](https://conventionalcommits.org) format | ✓ |
| `signoff` | Requires Signed-off-by line | ✓ |
| `signature` | Validates GPG/SSH signatures | ✓ |
| `identity` | Verifies committer identity | ✓ |
| `branchahead` | Limits commits ahead of main | ✓ |
| `commitbody` | Requires detailed commit body | ✗ |
| `jirareference` | Requires JIRA ticket references | ✗ |
| `spell` | Spell checking | ✗ |

## Output Formats

### Progressive Verbosity

```bash
# Minimal output for CI
gommitlint validate

# Technical details for developers
gommitlint validate -v

# Educational guidance for learning
gommitlint validate -vv
```

### Multiple Formats

```bash
# Human-readable (default)
gommitlint validate --format=text

# Machine-readable for automation
gommitlint validate --format=json

# CI/CD integration
gommitlint validate --format=github   # GitHub Actions
gommitlint validate --format=gitlab   # GitLab CI
```

## Exit Codes

- `0` • All validations passed
- `1` • Configuration or system error  
- `2` • Validation failures found

Perfect for automation and CI/CD pipelines.

## Documentation

- **[Usage Guide](docs/usage.md)** • Complete configuration and examples
- **[Architecture](docs/ARCHITECTURE.md)** • Design principles and implementation
- **[Configuration Examples](.gommitlint-full-example.yaml)** • All available options

## Contributing

We welcome contributions! See our [Contributing Guide](CONTRIBUTING.md) for details on:

- Reporting bugs
- Suggesting improvements  
- Submitting pull requests
- Testing and feedback

## License

- **Code:** [EUPL 1.2](LICENSE)
- **Documentation:** [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/)

---
