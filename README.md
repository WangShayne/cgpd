<p align="center">
  <h1 align="center">cgpd</h1>
  <p align="center">
    <strong>Create Git Push Docs</strong> — Generate commit messages and changelogs from staged changes using LLM
  </p>
  <p align="center">
    <a href="https://github.com/WangShayne/cgpd/actions/workflows/ci.yml"><img src="https://github.com/WangShayne/cgpd/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="https://github.com/WangShayne/cgpd/actions/workflows/release.yml"><img src="https://github.com/WangShayne/cgpd/actions/workflows/release.yml/badge.svg" alt="Release"></a>
    <a href="https://github.com/WangShayne/cgpd/releases"><img src="https://img.shields.io/github/v/release/WangShayne/cgpd" alt="GitHub release"></a>
    <a href="https://github.com/WangShayne/cgpd/blob/main/LICENSE"><img src="https://img.shields.io/github/license/WangShayne/cgpd" alt="License"></a>
  </p>
  <p align="center">
    <a href="README_zh.md">简体中文</a> | English
  </p>
</p>

---

## Features

- 🚀 Auto-generate concise commit messages from `git diff --staged`
- 📝 Generate detailed Markdown changelogs
- 🔧 Flexible configuration via file or environment variables
- 🌐 Compatible with OpenAI API and all compatible services
- 🌍 Multi-language output support (English / 简体中文)
- ⏳ Real-time progress indicator during LLM requests

## Quick Start

### Installation

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/WangShayne/cgpd/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/WangShayne/cgpd/main/install.ps1 | iex
```

### Other Installation Methods

<details>
<summary>Manual Download</summary>

Download the binary for your platform from the [Releases](https://github.com/WangShayne/cgpd/releases) page.

</details>

<details>
<summary>Build from Source</summary>

```bash
# Requires Go 1.22+
git clone https://github.com/WangShayne/cgpd.git
cd cgpd
go build -o cgpd .
```

</details>

<details>
<summary>Go Install</summary>

```bash
go install github.com/WangShayne/cgpd@latest
```

</details>

## Configuration

cgpd supports two configuration methods. Priority: **Environment Variables > Config File**.

### Config File

Config file search order:

1. Current directory: `./.config.yaml`
2. User home: `~/.cgpd/.config.yaml`

```yaml
llm:
  provider: "openai"              # or "openai-compatible"
  base_url: "https://api.openai.com"
  api_key: "sk-your-api-key-here"
  model: "gpt-4-turbo"
  language: "en"                  # en (English) or zh (简体中文)
```

**Global Configuration (Recommended):**

```bash
mkdir -p ~/.cgpd
cat > ~/.cgpd/.config.yaml << 'EOF'
llm:
  provider: "openai"
  base_url: "https://api.openai.com"
  api_key: "sk-your-api-key-here"
  model: "gpt-4-turbo"
  language: "en"
EOF
```

### Environment Variables

```bash
export CGPD_LLM_PROVIDER="openai"
export OPENAI_API_KEY="sk-your-api-key-here"
export CGPD_LLM_MODEL="gpt-4-turbo"

# Optional (defaults to https://api.openai.com)
export CGPD_LLM_BASE_URL="https://api.openai.com"
export CGPD_LANGUAGE="en"
```

<details>
<summary>All Supported Environment Variables</summary>

| Config Key       | Environment Variables (by priority)                         |
| ---------------- | ----------------------------------------------------------- |
| `llm.provider`   | `CGPD_LLM_PROVIDER`, `LLM_PROVIDER`                         |
| `llm.base_url`   | `CGPD_LLM_BASE_URL`, `LLM_BASE_URL`, `OPENAI_BASE_URL`      |
| `llm.api_key`    | `CGPD_LLM_API_KEY`, `OPENAI_API_KEY`, `LLM_API_KEY`         |
| `llm.model`      | `CGPD_LLM_MODEL`, `LLM_MODEL`, `OPENAI_MODEL`               |
| `llm.language`   | `CGPD_LANGUAGE`, `CGPD_LLM_LANGUAGE`                        |

</details>

<details>
<summary>Third-Party API Examples</summary>

```yaml
# Azure OpenAI
llm:
  provider: "openai-compatible"
  base_url: "https://your-resource.openai.azure.com/openai/deployments/your-deployment"
  api_key: "your-azure-api-key"
  model: "gpt-4"

# Google Gemini (OpenAI-compatible endpoint)
llm:
  provider: "openai-compatible"
  base_url: "https://generativelanguage.googleapis.com/v1beta/openai"
  api_key: "your-gemini-api-key"
  model: "gemini-2.5-flash"

# DeepSeek
llm:
  provider: "openai-compatible"
  base_url: "https://api.deepseek.com/v1"
  api_key: "your-api-key"
  model: "deepseek-chat"

# Ollama (Local)
llm:
  provider: "openai-compatible"
  base_url: "http://localhost:11434/v1"
  api_key: "ollama"
  model: "llama3"
```

</details>

## Usage

### Generate Commit Message (Default)

```bash
git add .
cgpd
# Output: Add user authentication with JWT tokens

# Use directly with git commit
git commit -m "$(cgpd)"
```

### Generate Changelog

```bash
cgpd --docs
# Output: docs/history/2025-12-26-143052.md
```

<details>
<summary>Example Generated Changelog</summary>

```markdown
# Changelog

## Summary

This update adds user authentication using JWT tokens.

## Changes

### API
- Add `/api/auth/login` endpoint
- Add `/api/auth/refresh` token refresh endpoint

### Configuration
- Add `JWT_SECRET` environment variable
- Add `TOKEN_EXPIRY` configuration

## Migration Notes

Configure `JWT_SECRET` in environment variables before starting the service.

## Changed Files

- `internal/auth/jwt.go`
- `internal/api/handlers.go`
- `config/config.go`
```

</details>

### CLI Options

```
Usage:
  cgpd [flags]

Flags:
      --docs      Generate detailed Markdown changelog
  -h, --help      Show help
  -v, --version   Show version
```

## Workflow Examples

### Daily Development

```bash
git add .
git commit -m "$(cgpd)"
git push
```

### Release

```bash
git add .
cgpd --docs
git add docs/history/
git commit -m "$(cgpd)"
git tag v1.0.0
git push --tags
```

### Git Hooks Integration

Create `.git/hooks/prepare-commit-msg`:

```bash
#!/bin/bash
if [ -z "$(cat $1)" ]; then
  cgpd > $1
fi
```

```bash
chmod +x .git/hooks/prepare-commit-msg
```

## FAQ

<details>
<summary>Q: "no staged changes found"</summary>

Stage your changes first:

```bash
git add .
# or specific files
git add src/main.go
```

</details>

<details>
<summary>Q: "config not found"</summary>

Create a config file or set environment variables:

```bash
# Option 1: Global config
mkdir -p ~/.cgpd
cat > ~/.cgpd/.config.yaml << 'EOF'
llm:
  provider: "openai"
  api_key: "sk-xxx"
  model: "gpt-4-turbo"
EOF

# Option 2: Environment variables
export CGPD_LLM_PROVIDER="openai"
export OPENAI_API_KEY="sk-xxx"
export CGPD_LLM_MODEL="gpt-4-turbo"
```

</details>

<details>
<summary>Q: How to use local models?</summary>

Use Ollama or similar local services:

```yaml
llm:
  provider: "openai-compatible"
  base_url: "http://localhost:11434/v1"
  api_key: "ollama"
  model: "llama3"
```

</details>

<details>
<summary>Q: How to uninstall?</summary>

**Linux / macOS:**

```bash
sudo rm /usr/local/bin/cgpd
```

**Windows:**

```powershell
Remove-Item "$env:LOCALAPPDATA\Programs\cgpd" -Recurse -Force
```

</details>

## Project Structure

```
cgpd/
├── main.go
├── cmd/
│   └── root.go
├── internal/
│   ├── config/
│   ├── git/
│   ├── llm/
│   └── spinner/
├── .github/workflows/
├── install.sh
├── install.ps1
└── README.md
```

## Contributing

```bash
git clone https://github.com/WangShayne/cgpd.git
cd cgpd
go mod download
go build -o cgpd .
go test ./...
```

## License

[MIT License](LICENSE)
