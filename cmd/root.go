package cmd

import (
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

	"github.com/spf13/cobra"
)

var flagDocs bool

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

	diff, err := git.StagedDiff(ctx)
	if err != nil {
		return err
	}

	client, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return err
	}

	if flagDocs {
		markdown, err := client.GenerateDocs(ctx, diff)
		if err != nil {
			return err
		}
		if strings.TrimSpace(markdown) == "" {
			return errors.New("LLM returned empty docs")
		}
		path, err := writeDocsFile(markdown)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), path)
		return nil
	}

	msg, err := client.GenerateCommitMessage(ctx, diff)
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
