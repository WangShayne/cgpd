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
	flagDocs           bool
	flagSkipSecurity   bool
	flagNonInteractive bool
	flagAssumeYes      bool
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

	ignoreRules, err := security.LoadIgnoreFile(cfg.Security.IgnoreFile)
	if err != nil {
		return fmt.Errorf("load ignore file: %w", err)
	}

	// Get staged files first for both security check and docs
	files, err := git.StagedFiles(ctx)
	if err != nil {
		return err
	}

	// Check for sensitive files unless skipped
	if !flagSkipSecurity && cfg.Security.Enabled {
		filesForCheck := security.FilterPaths(files, ignoreRules)
		sensitiveFiles, err := security.CheckSensitiveFiles(cfg.Security, filesForCheck)

		if err != nil {
			return err
		}

		if len(sensitiveFiles) > 0 {
			nonInteractive := flagNonInteractive || os.Getenv("CGPD_NON_INTERACTIVE") == "1"
			assumeYes := flagAssumeYes || os.Getenv("CGPD_YES") == "1"

			ok, err := confirmSensitiveFiles(sensitiveFiles, cfg.LLM.Language, nonInteractive, assumeYes)
			if err != nil {
				return err
			}
			if !ok {
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

	if !flagSkipSecurity && cfg.Security.Enabled && cfg.Security.ScanDiff {
		nonInteractive := flagNonInteractive || os.Getenv("CGPD_NON_INTERACTIVE") == "1"
		assumeYes := flagAssumeYes || os.Getenv("CGPD_YES") == "1"

		findings := security.FilterFindings(security.ScanStagedDiff(diff), ignoreRules)
		if len(findings) > 0 {
			ok, err := confirmDiffFindings(findings, cfg.LLM.Language, nonInteractive, assumeYes)
			if err != nil {
				return err
			}
			if !ok {
				if cfg.LLM.Language == "zh" {
					return errors.New("操作已取消")
				}
				return errors.New("operation cancelled by user")
			}
		}
	}

	llmDiff := diff
	const maxLLMDiffBytes = 300000
	const maxLLMDiffLines = 6000
	if len(llmDiff) > maxLLMDiffBytes || strings.Count(llmDiff, "\n") > maxLLMDiffLines {
		stat, err := git.StagedDiffStat(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "cgpd: staged diff too large (%d bytes); using 'git diff --staged --stat' summary\n", len(llmDiff))
		llmDiff = "NOTE: staged diff too large; using --stat summary.\n\n" + stat
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
		markdown, err := client.GenerateDocs(ctx, llmDiff)
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
	msg, err := client.GenerateCommitMessage(ctx, llmDiff)
	spin.Stop()
	if err != nil {
		return err
	}
	msg = normalizeCommitSubject(msg)
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
	rootCmd.Flags().BoolVar(&flagNonInteractive, "non-interactive", false, "Fail instead of prompting when checks need confirmation")
	rootCmd.Flags().BoolVar(&flagAssumeYes, "yes", false, "Assume yes for prompts (use with care)")
}

func confirmSensitiveFiles(files []string, lang string, nonInteractive bool, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}

	if nonInteractive {
		joined := strings.Join(files, ", ")
		if lang == "zh" {
			return false, fmt.Errorf("非交互模式下检测到敏感文件: %s（使用 --yes 继续，或使用 --skip-security 跳过）", joined)
		}
		return false, fmt.Errorf("sensitive files detected in non-interactive mode: %s (use --yes to continue or --skip-security)", joined)
	}

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
		return false, fmt.Errorf("read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func confirmDiffFindings(findings []security.Finding, lang string, nonInteractive bool, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}

	if nonInteractive {
		if lang == "zh" {
			return false, errors.New("非交互模式下检测到疑似密钥/敏感内容（使用 --yes 继续，或使用 --skip-security 跳过）")
		}
		return false, errors.New("potential secrets detected in non-interactive mode (use --yes to continue or --skip-security)")
	}

	var msg, prompt string
	if lang == "zh" {
		msg = "\n⚠️  检测到以下疑似密钥/敏感内容（来自暂存 diff 的新增行）:\n"
		prompt = "\n确认继续? (y/N): "
	} else {
		msg = "\n⚠️  Potential secrets detected (from added staged diff lines):\n"
		prompt = "\nContinue anyway? (y/N): "
	}

	fmt.Fprint(os.Stderr, msg)
	for _, f := range findings {
		loc := f.File
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		fp := f.Fingerprint
		if len(fp) > 12 {
			fp = fp[:12]
		}
		fmt.Fprintf(os.Stderr, "  - %s (%s) %s [%s]\n", loc, f.RuleID, f.Preview, fp)
	}
	fmt.Fprint(os.Stderr, prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
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

func normalizeCommitSubject(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	if i := strings.IndexByte(s, '\n'); i != -1 {
		s = s[:i]
	}

	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"'`")
	s = strings.TrimSpace(s)

	s = strings.TrimPrefix(s, "- ")
	s = strings.TrimPrefix(s, "* ")
	s = strings.TrimPrefix(s, "# ")

	s = strings.Join(strings.Fields(s), " ")
	rs := []rune(s)
	if len(rs) <= 72 {
		return s
	}
	return truncateRunes(s, 72)
}

func truncateRunes(s string, max int) string {
	rs := []rune(s)
	if max <= 0 {
		return ""
	}
	if len(rs) <= max {
		return s
	}
	if max <= 3 {
		return string(rs[:max])
	}
	return string(rs[:max-3]) + "..."
}
