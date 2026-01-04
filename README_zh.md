<p align="center">
  <h1 align="center">cgpd</h1>
  <p align="center">
    <strong>Create Git Push Docs</strong> — 使用 LLM 从暂存变更生成 commit 信息和变更文档
  </p>
  <p align="center">
    <a href="https://github.com/WangShayne/cgpd/actions/workflows/ci.yml"><img src="https://github.com/WangShayne/cgpd/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="https://github.com/WangShayne/cgpd/actions/workflows/release.yml"><img src="https://github.com/WangShayne/cgpd/actions/workflows/release.yml/badge.svg" alt="Release"></a>
    <a href="https://github.com/WangShayne/cgpd/releases"><img src="https://img.shields.io/github/v/release/WangShayne/cgpd" alt="GitHub release"></a>
    <a href="https://github.com/WangShayne/cgpd/blob/main/LICENSE"><img src="https://img.shields.io/github/license/WangShayne/cgpd" alt="License"></a>
  </p>
  <p align="center">
    简体中文 | <a href="README.md">English</a>
  </p>
</p>

---

## 功能特性

- 🚀 根据 `git diff --staged` 自动生成简洁的 commit 信息
- 📝 生成详细的 Markdown 格式变更文档
- 🔧 支持配置文件和环境变量两种配置方式
- 🌐 兼容 OpenAI API 及所有兼容接口
- 🌍 支持多语言输出（English / 简体中文）
- ⏳ LLM 请求时显示实时进度

## 快速开始

### 安装

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/WangShayne/cgpd/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/WangShayne/cgpd/main/install.ps1 | iex
```

### 其他安装方式

<details>
<summary>手动下载</summary>

从 [Releases](https://github.com/WangShayne/cgpd/releases) 页面下载对应平台的二进制文件。

</details>

<details>
<summary>从源码构建</summary>

```bash
# 需要 Go 1.22+
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

## 配置

cgpd 支持两种配置方式，优先级：**环境变量 > 配置文件**。

### 配置文件

配置文件搜索顺序：

1. 当前目录：`./.config.yaml`
2. 用户目录：`~/.cgpd/.config.yaml`

```yaml
llm:
  provider: "openai"              # 或 "openai-compatible"
  base_url: "https://api.openai.com"
  api_key: "sk-your-api-key-here"
  model: "gpt-4-turbo"
  language: "zh"                  # en (English) 或 zh (简体中文)
```

**全局配置（推荐）：**

```bash
mkdir -p ~/.cgpd
cat > ~/.cgpd/.config.yaml << 'EOF'
llm:
  provider: "openai"
  base_url: "https://api.openai.com"
  api_key: "sk-your-api-key-here"
  model: "gpt-4-turbo"
  language: "zh"
EOF
```

### 环境变量

```bash
export CGPD_LLM_PROVIDER="openai"
export OPENAI_API_KEY="sk-your-api-key-here"
export CGPD_LLM_MODEL="gpt-4-turbo"

# 可选（默认为 https://api.openai.com）
export CGPD_LLM_BASE_URL="https://api.openai.com"
export CGPD_LANGUAGE="zh"
```

<details>
<summary>支持的环境变量</summary>

| 配置项           | 环境变量（按优先级）                                    |
| ---------------- | ------------------------------------------------------- |
| `llm.provider`   | `CGPD_LLM_PROVIDER`, `LLM_PROVIDER`                     |
| `llm.base_url`   | `CGPD_LLM_BASE_URL`, `LLM_BASE_URL`, `OPENAI_BASE_URL`  |
| `llm.api_key`    | `CGPD_LLM_API_KEY`, `OPENAI_API_KEY`, `LLM_API_KEY`     |
| `llm.model`      | `CGPD_LLM_MODEL`, `LLM_MODEL`, `OPENAI_MODEL`           |
| `llm.language`   | `CGPD_LANGUAGE`, `CGPD_LLM_LANGUAGE`                    |

</details>

<details>
<summary>第三方 API 配置示例</summary>

```yaml
# Azure OpenAI
llm:
  provider: "openai-compatible"
  base_url: "https://your-resource.openai.azure.com/openai/deployments/your-deployment"
  api_key: "your-azure-api-key"
  model: "gpt-4"

# Google Gemini (OpenAI 兼容端点)
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

# Ollama 本地模型
llm:
  provider: "openai-compatible"
  base_url: "http://localhost:11434/v1"
  api_key: "ollama"
  model: "llama3"
```

</details>

## 使用方法

### 生成 Commit 信息（默认）

```bash
git add .
cgpd
# 输出：添加用户认证功能，使用 JWT 令牌

# 直接用于 git commit
git commit -m "$(cgpd)"
```

### 生成变更文档

```bash
cgpd --docs
# 输出：docs/history/2025-12-26-143052.md
```

### 管理 Git Hooks

```bash
# 安装 Git hook 实现自动生成提交信息
cgpd hooks install

# 查看 hook 状态
cgpd hooks list

# 卸载 hook
cgpd hooks uninstall
```

安装 hook 后，运行 `git commit`（不使用 `-m`）时将自动生成提交信息。

<details>
<summary>生成的文档示例</summary>

```markdown
# 变更日志

## 概述

本次更新添加了用户认证功能，使用 JWT 令牌进行身份验证。

## 详细变更

### API
- 添加 `/api/auth/login` 登录接口
- 添加 `/api/auth/refresh` 令牌刷新接口

### 配置
- 新增 `JWT_SECRET` 环境变量配置
- 新增 `TOKEN_EXPIRY` 过期时间配置

## 迁移说明

需要在环境变量中配置 `JWT_SECRET`，否则服务将无法启动。

## 变更文件

- `internal/auth/jwt.go`
- `internal/api/handlers.go`
- `config/config.go`
```

</details>

### 命令行选项

```text
Usage:
  cgpd [flags]

Flags:
      --docs      生成详细的 Markdown 变更文档
  -h, --help      显示帮助信息
  -v, --version   显示版本信息
```

## 工作流示例

### 日常开发

```bash
git add .
git commit -m "$(cgpd)"
git push
```

### 版本发布

```bash
git add .
cgpd --docs
git add docs/history/
git commit -m "$(cgpd)"
git tag v1.0.0
git push --tags
```

### Git Hooks 集成

**自动安装（推荐）：**

```bash
# 安装 Git hook 实现自动生成提交信息
cgpd hooks install

# 查看 hook 状态
cgpd hooks list

# 卸载 hook
cgpd hooks uninstall
```

安装后，当您运行 `git commit` 而不使用 `-m` 标志时，将自动生成提交信息。

**手动安装：**

创建 `.git/hooks/prepare-commit-msg`：

```bash
#!/bin/bash
if [ -z "$(cat $1)" ]; then
  cgpd > $1
fi
```

```bash
chmod +x .git/hooks/prepare-commit-msg
```

## 常见问题

<details>
<summary>Q: 提示 "no staged changes found"</summary>

请先暂存变更：

```bash
git add .
# 或指定文件
git add src/main.go
```

</details>

<details>
<summary>Q: 提示 "config not found"</summary>

创建配置文件或设置环境变量：

```bash
# 方式一：全局配置
mkdir -p ~/.cgpd
cat > ~/.cgpd/.config.yaml << 'EOF'
llm:
  provider: "openai"
  api_key: "sk-xxx"
  model: "gpt-4-turbo"
EOF

# 方式二：环境变量
export CGPD_LLM_PROVIDER="openai"
export OPENAI_API_KEY="sk-xxx"
export CGPD_LLM_MODEL="gpt-4-turbo"
```

</details>

<details>
<summary>Q: 如何使用本地模型？</summary>

使用 Ollama 等本地服务：

```yaml
llm:
  provider: "openai-compatible"
  base_url: "http://localhost:11434/v1"
  api_key: "ollama"
  model: "llama3"
```

</details>

<details>
<summary>Q: 如何卸载？</summary>

**Linux / macOS:**

```bash
sudo rm /usr/local/bin/cgpd
```

**Windows:**

```powershell
Remove-Item "$env:LOCALAPPDATA\Programs\cgpd" -Recurse -Force
```

</details>

## 项目结构

```text
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

## 参与开发

```bash
git clone https://github.com/WangShayne/cgpd.git
cd cgpd
go mod download
go build -o cgpd .
go test ./...
```

## 许可证

[MIT License](LICENSE)
