package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cgpd/internal/config"
	"cgpd/internal/git"
	"cgpd/internal/security"

	"github.com/spf13/cobra"
)

var (
	flagBaselineFile string
	flagScanDiff     bool
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Security helpers (scan/baseline)",
	Long:  "Scan staged changes for potential secrets and manage a baseline file.",
}

var baselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Manage baseline for secret findings",
}

var baselineWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Write baseline file from current staged findings",
	RunE:  runBaselineWrite,
}

var baselineCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Fail if new findings appear compared to baseline",
	RunE:  runBaselineCheck,
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan staged changes for sensitive files/secrets",
	RunE:  runSecurityScan,
}

func init() {
	rootCmd.AddCommand(securityCmd)
	securityCmd.AddCommand(scanCmd)
	securityCmd.AddCommand(baselineCmd)
	baselineCmd.AddCommand(baselineWriteCmd)
	baselineCmd.AddCommand(baselineCheckCmd)

	scanCmd.Flags().BoolVar(&flagScanDiff, "diff", false, "Scan staged diff added lines for potential secrets")
	baselineCmd.PersistentFlags().StringVar(&flagBaselineFile, "baseline-file", "", "Override baseline file path")
}

func runBaselineWrite(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ignoreRules, err := security.LoadIgnoreFile(cfg.Security.IgnoreFile)
	if err != nil {
		return fmt.Errorf("load ignore file: %w", err)
	}

	diff, err := git.StagedDiff(ctx)
	if err != nil {
		return err
	}

	findings := security.FilterFindings(security.ScanStagedDiff(diff), ignoreRules)

	path := cfg.Security.BaselineFile
	if flagBaselineFile != "" {
		path = flagBaselineFile
	}
	if path == "" {
		return errors.New("baseline file path is empty")
	}

	if err := security.WriteBaseline(path, findings); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), path)
	return nil
}

func runBaselineCheck(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ignoreRules, err := security.LoadIgnoreFile(cfg.Security.IgnoreFile)
	if err != nil {
		return fmt.Errorf("load ignore file: %w", err)
	}

	path := cfg.Security.BaselineFile
	if flagBaselineFile != "" {
		path = flagBaselineFile
	}
	if path == "" {
		return errors.New("baseline file path is empty")
	}

	baseline, err := security.LoadBaseline(path)
	if err != nil {
		return err
	}

	diff, err := git.StagedDiff(ctx)
	if err != nil {
		return err
	}

	findings := security.FilterFindings(security.ScanStagedDiff(diff), ignoreRules)
	newOnes := security.NewFindingsSinceBaseline(findings, baseline)
	if len(newOnes) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "OK")
		return nil
	}

	lang := cfg.LLM.Language
	if lang == "zh" {
		fmt.Fprintln(os.Stderr, "检测到新的疑似密钥/敏感内容（相对 baseline）:")
	} else {
		fmt.Fprintln(os.Stderr, "New potential secrets detected (vs baseline):")
	}
	for _, f := range newOnes {
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

	if lang == "zh" {
		return errors.New("baseline 检查失败")
	}
	return errors.New("baseline check failed")
}

func runSecurityScan(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ignoreRules, err := security.LoadIgnoreFile(cfg.Security.IgnoreFile)
	if err != nil {
		return fmt.Errorf("load ignore file: %w", err)
	}

	files, err := git.StagedFiles(ctx)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "OK")
		return nil
	}

	filesForCheck := security.FilterPaths(files, ignoreRules)
	sensitiveFiles, err := security.CheckSensitiveFiles(cfg.Security, filesForCheck)
	if err != nil {
		return err
	}

	var findings []security.Finding
	if cfg.Security.ScanDiff || flagScanDiff {
		diff, err := git.StagedDiff(ctx)
		if err != nil {
			return err
		}
		findings = security.FilterFindings(security.ScanStagedDiff(diff), ignoreRules)
	}

	if len(sensitiveFiles) == 0 && len(findings) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "OK")
		return nil
	}

	lang := cfg.LLM.Language
	if len(sensitiveFiles) > 0 {
		if lang == "zh" {
			fmt.Fprintln(os.Stderr, "检测到敏感文件:")
		} else {
			fmt.Fprintln(os.Stderr, "Sensitive files detected:")
		}
		for _, f := range sensitiveFiles {
			fmt.Fprintf(os.Stderr, "  - %s\n", f)
		}
	}

	if len(findings) > 0 {
		if lang == "zh" {
			fmt.Fprintln(os.Stderr, "检测到疑似密钥/敏感内容（来自暂存 diff 新增行）:")
		} else {
			fmt.Fprintln(os.Stderr, "Potential secrets detected (from added staged diff lines):")
		}
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
	}

	if lang == "zh" {
		return errors.New("安全扫描失败")
	}
	return errors.New("security scan failed")
}
