package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func StagedDiff(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--staged")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff --staged: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	diff := strings.TrimSpace(string(out))
	if diff == "" {
		return "", errors.New("no staged changes found (use 'git add' first)")
	}
	return diff + "\n", nil
}

func StagedFiles(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--staged", "--name-only")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff --staged --name-only: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		if f := strings.TrimSpace(line); f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}
