# cgpd

> Create Git Push Docs - 使用 LLM 从暂存变更生成 commit 信息和变更文档

[![CI](https://github.com/YOUR_USERNAME/cgpd/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/cgpd/actions/workflows/ci.yml)
[![Release](https://github.com/YOUR_USERNAME/cgpd/actions/workflows/release.yml/badge.svg)](https://github.com/YOUR_USERNAME/cgpd/actions/workflows/release.yml)

## 功能特性

- 🚀 根据 `git diff --staged` 自动生成简洁的 commit 信息
- 📝 生成详细的 Markdown 格式变更文档
- 🔧 支持配置文件和环境变量两种配置方式
- 🌐 兼容 OpenAI API 及所有兼容接口

## 安装

### 从 Release 下载

从 [Releases](https://github.com/YOUR_USERNAME/cgpd/releases) 页面下载对应平台的二进制文件。

**Linux / macOS:**

```bash
# Linux amd64
curl -LO https://github.com/YOUR_USERNAME/cgpd/releases/latest/download/cgpd-linux-amd64
chmod +x cgpd-linux-amd64
sudo mv cgpd-linux-amd64 /usr/local/bin/cgpd

# macOS arm64 (Apple Silicon)
curl -LO https://github.com/YOUR_USERNAME/cgpd/releases/latest/download/cgpd-darwin-arm64
chmod +x cgpd-darwin-arm64
sudo mv cgpd-darwin-arm64 /usr/local/bin/cgpd

# macOS amd64 (Intel)
curl -LO https://github.com/YOUR_USERNAME/cgpd/releases/latest/download/cgpd-darwin-amd64
chmod +x cgpd-darwin-amd64
sudo mv cgpd-darwin-amd64 /usr/local/bin/cgpd
```

**Windows (PowerShell):**

```powershell
# 下载
Invoke-WebRequest -Uri "https://github.com/YOUR_USERNAME/cgpd/releases/latest/download/cgpd-windows-amd64.exe" -OutFile "cgpd.exe"

# 移动到 PATH 目录（以管理员身份运行）
Move-Item cgpd.exe C:\Windows\System32\
```

### 从源码构建

```bash
# 需要 Go 1.22+
git clone https://github.com/YOUR_USERNAME/cgpd.git
cd cgpd
go build -o cgpd .
```

### 使用 Go Install

```bash
go install github.com/YOUR_USERNAME/cgpd@latest
```

## 配置

cgpd 支持两种配置方式，优先级：环境变量 > 配置文件。

### 方式一：配置文件

在项目根目录（或任意父目录）创建 `.cgpd.yaml`：

```yaml
llm:
  provider: "openai"              # 或 "openai-compatible"
  base_url: "https://api.openai.com"
  api_key: "sk-your-api-key-here"
  model: "gpt-4-turbo"
```

### 方式二：环境变量

```bash
# 必需
export CGPD_LLM_PROVIDER="openai"
export OPENAI_API_KEY="sk-your-api-key-here"
export CGPD_LLM_MODEL="gpt-4-turbo"

# 可选（默认为 https://api.openai.com）
export CGPD_LLM_BASE_URL="https://api.openai.com"
```

### 支持的环境变量

| 配置项           | 环境变量（按优先级）                                    |
| ---------------- | ------------------------------------------------------- |
| `llm.provider`   | `CGPD_LLM_PROVIDER`, `LLM_PROVIDER`                     |
| `llm.base_url`   | `CGPD_LLM_BASE_URL`, `LLM_BASE_URL`, `OPENAI_BASE_URL`  |
| `llm.api_key`    | `CGPD_LLM_API_KEY`, `OPENAI_API_KEY`, `LLM_API_KEY`     |
| `llm.model`      | `CGPD_LLM_MODEL`, `LLM_MODEL`, `OPENAI_MODEL`           |

### 使用第三方 API

支持任何 OpenAI 兼容的 API 服务：

```yaml
# 使用 Azure OpenAI
llm:
  provider: "openai-compatible"
  base_url: "https://your-resource.openai.azure.com"
  api_key: "your-azure-api-key"
  model: "gpt-4"

# 使用 Ollama 本地模型
llm:
  provider: "openai-compatible"
  base_url: "http://localhost:11434/v1"
  api_key: "ollama"
  model: "llama3"

# 使用其他兼容服务
llm:
  provider: "openai-compatible"
  base_url: "https://api.deepseek.com"
  api_key: "your-api-key"
  model: "deepseek-chat"
```

## 使用方法

### 生成 Commit 信息（默认模式）

```bash
# 1. 暂存你的更改
git add .

# 2. 生成 commit 信息（输出到 stdout）
cgpd

# 3. 直接用于 git commit
git commit -m "$(cgpd)"
```

**输出示例：**

```
Add user authentication with JWT tokens
```

### 生成变更文档

```bash
cgpd --docs
```

**输出：**

```
docs/history/2025-12-26-143052.md
```

**生成的文档示例：**

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
```

### 命令行选项

```
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
# 编写代码后
git add .
git commit -m "$(cgpd)"
git push
```

### 版本发布

```bash
# 生成变更文档
git add .
cgpd --docs

# 将文档也加入 commit
git add docs/history/
git commit -m "$(cgpd)"
git tag v1.0.0
git push --tags
```

### 与 Git Hooks 集成

创建 `.git/hooks/prepare-commit-msg`：

```bash
#!/bin/bash
# 如果没有提供 commit 信息，使用 cgpd 生成
if [ -z "$(cat $1)" ]; then
  cgpd > $1
fi
```

```bash
chmod +x .git/hooks/prepare-commit-msg
```

## 项目结构

```
cgpd/
├── main.go                      # 程序入口
├── cmd/
│   └── root.go                  # CLI 命令定义
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载
│   ├── git/
│   │   └── git.go               # Git 操作
│   └── llm/
│       └── client.go            # LLM 客户端
├── docs/
│   └── history/                 # 变更文档目录
├── .github/
│   └── workflows/
│       ├── ci.yml               # CI 工作流
│       └── release.yml          # 发布工作流
└── README.md
```

## 开发

```bash
# 克隆仓库
git clone https://github.com/YOUR_USERNAME/cgpd.git
cd cgpd

# 安装依赖
go mod download

# 构建
go build -o cgpd .

# 运行测试
go test ./...
```

## 发布新版本

```bash
# 创建标签触发自动构建
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions 将自动：

1. 构建 Linux/macOS/Windows 多平台二进制文件
2. 生成 SHA256 校验和
3. 创建 GitHub Release

## 常见问题

### Q: 提示 "no staged changes found"

确保使用 `git add` 暂存了变更：

```bash
git add .
# 或指定文件
git add src/main.go
```

### Q: 提示 "config file .cgpd.yaml not found"

创建配置文件或设置环境变量：

```bash
export CGPD_LLM_PROVIDER="openai"
export OPENAI_API_KEY="sk-xxx"
export CGPD_LLM_MODEL="gpt-4-turbo"
```

### Q: 如何使用本地模型？

使用 Ollama 等本地服务：

```yaml
llm:
  provider: "openai-compatible"
  base_url: "http://localhost:11434/v1"
  api_key: "ollama"
  model: "llama3"
```

## License

MIT License
