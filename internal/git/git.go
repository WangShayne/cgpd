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
