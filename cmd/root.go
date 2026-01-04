package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cgpd/internal/config"
	"cgpd/internal/git"
	"cgpd/internal/llm"
	"cgpd/internal/security"
	"cgpd/internal/spinner"

	"github.com/spf13/cobra"
)

var (
	flagDocs         bool
	flagSkipSecurity bool
)

var rootCmd = &cobra.Command{
	Use:          "cgpd",
	Short:        "Generate commit messages or changelogs from staged changes using LLM",
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	RunE:         run,
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func run(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 90*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Get staged files first for both security check and docs
	files, err := git.StagedFiles(ctx)
	if err != nil {
		return err
	}

	// Check for sensitive files unless skipped
	if !flagSkipSecurity && cfg.Security.Enabled {
		sensitiveFiles, err := security.CheckSensitiveFiles(cfg.Security, files)
		if err != nil {
			return err
		}

		if len(sensitiveFiles) > 0 {
			if !confirmSensitiveFiles(sensitiveFiles, cfg.LLM.Language) {
				if cfg.LLM.Language == "zh" {
					return errors.New("操作已取消")
				}
				return errors.New("operation cancelled by user")
			}
		}
	}

	diff, err := git.StagedDiff(ctx)
	if err != nil {
		return err
	}

	client, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return err
	}

	if flagDocs {
		spinMsg := "Generating documentation..."
		if cfg.LLM.Language == "zh" {
			spinMsg = "正在生成文档..."
		}
		spin := spinner.New(os.Stderr, spinMsg)
		spin.Start()
		markdown, err := client.GenerateDocs(ctx, diff)
		spin.Stop()
		if err != nil {
			return err
		}
		if strings.TrimSpace(markdown) == "" {
			return errors.New("LLM returned empty docs")
		}
		markdown = appendFilesSection(markdown, files, cfg.LLM.Language)
		path, err := writeDocsFile(markdown)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), path)
		return nil
	}

	spinMsg := "Generating commit message..."
	if cfg.LLM.Language == "zh" {
		spinMsg = "正在生成提交信息..."
	}
	spin := spinner.New(os.Stderr, spinMsg)
	spin.Start()
	msg, err := client.GenerateCommitMessage(ctx, diff)
	spin.Stop()
	if err != nil {
		return err
	}
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return errors.New("LLM returned an empty commit message")
	}
	fmt.Fprintln(cmd.OutOrStdout(), msg)
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&flagDocs, "docs", false, "Generate detailed Markdown changelog")
	rootCmd.Flags().BoolVar(&flagSkipSecurity, "skip-security", false, "Skip sensitive file detection")
}

func confirmSensitiveFiles(files []string, lang string) bool {
	var msg, prompt string

	if lang == "zh" {
		msg = "\n⚠️  检测到以下敏感文件将被提交:\n"
		prompt = "\n确认继续提交? (y/N): "
	} else {
		msg = "\n⚠️  Sensitive files detected:\n"
		prompt = "\nContinue anyway? (y/N): "
	}

	fmt.Fprint(os.Stderr, msg)
	for _, f := range files {
		fmt.Fprintf(os.Stderr, "  - %s\n", f)
	}
	fmt.Fprint(os.Stderr, prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}


func writeDocsFile(content string) (string, error) {
	dir := filepath.Join(".", "docs", "history")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create docs directory: %w", err)
	}

	name := time.Now().Format("2006-01-02-150405") + ".md"
	path := filepath.Join(dir, name)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("create history file: %w", err)
	}
	defer f.Close()

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if _, err := f.WriteString(content); err != nil {
		return "", fmt.Errorf("write history file: %w", err)
	}

	return path, nil
}

func appendFilesSection(markdown string, files []string, lang string) string {
	if len(files) == 0 {
		return markdown
	}

	title := "## Changed Files\n\n"
	if strings.ToLower(strings.TrimSpace(lang)) == "zh" {
		title = "## 变更文件\n\n"
	}

	var sb strings.Builder
	sb.WriteString(strings.TrimRight(markdown, "\n"))
	sb.WriteString("\n\n")
	sb.WriteString(title)
	for _, f := range files {
		sb.WriteString("- `")
		sb.WriteString(f)
		sb.WriteString("`\n")
	}
	return sb.String()
}
