package security

import (
	"path/filepath"
	"strings"

	"cgpd/internal/config"
)

// defaultSensitivePatterns 默认的敏感文件模式列表
var defaultSensitivePatterns = []string{
	".env",
	".env.*",
	"*.pem",
	"*.key",
	"*_rsa",
	"*.p12",
	"*.pfx",
	"*secret*",
	"*password*",
	"*credentials*",
	".config.yaml",
	".cgpd.yaml",
	"id_rsa",
	"id_dsa",
	"id_ecdsa",
	"id_ed25519",
	"*.keystore",
	"*.jks",
	"*token*",
	"*apikey*",
	"*api_key*",
	"*private*key*",
}

// CheckSensitiveFiles 检查文件列表中是否包含敏感文件
// cfg: 安全配置,包含自定义模式和排除规则
// files: 待检查的文件列表
// 返回: (sensitiveFiles []string, err error)
func CheckSensitiveFiles(cfg config.SecurityConfig, files []string) ([]string, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// 合并默认模式和用户自定义模式
	patterns := append([]string{}, defaultSensitivePatterns...)
	patterns = append(patterns, cfg.AdditionalPatterns...)

	var sensitiveFiles []string
	for _, file := range files {
		if isSensitiveFile(file, patterns, cfg.ExcludePatterns) {
			sensitiveFiles = append(sensitiveFiles, file)
		}
	}

	return sensitiveFiles, nil
}

// isSensitiveFile 检查单个文件是否为敏感文件
// filename: 完整文件路径
// patterns: 敏感文件模式列表
// excludePatterns: 排除模式列表
// 返回: true 表示该文件是敏感文件
func isSensitiveFile(filename string, patterns, excludePatterns []string) bool {
	// 提取文件名(移除路径)
	basename := filepath.Base(filename)
	lowerBasename := strings.ToLower(basename)

	// 先检查是否在排除列表中
	for _, pattern := range excludePatterns {
		if matchPattern(lowerBasename, strings.ToLower(pattern)) {
			return false
		}
	}

	// 检查是否匹配敏感文件模式
	for _, pattern := range patterns {
		if matchPattern(lowerBasename, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// matchPattern 执行模式匹配(不区分大小写)
func matchPattern(filename, pattern string) bool {
	matched, err := filepath.Match(pattern, filename)
	if err != nil {
		// 模式无效,忽略该模式
		return false
	}
	return matched
}
