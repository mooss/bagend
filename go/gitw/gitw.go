// Package gitw provides git wrappers.
package gitw

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileIsClean requires the file to be tracked and inside the current Git repository, and rejects
// files with unstaged changes.
func FileIsClean(path string) error {
	ctx := context.Background()
	repo, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return errors.New("cannot execute outside a git repository")
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve %q: %w", path, err)
	}

	root := strings.TrimSpace(string(repo))
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		return err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%q is outside the git repository %q", path, root)
	}

	tracked, err := exec.CommandContext(
		ctx, "git", "--literal-pathspecs", "-C", root, "ls-files", "--", relative,
	).Output()
	if err != nil {
		return fmt.Errorf("checking git tracking of %q: %w", path, err)
	}
	if strings.TrimSpace(string(tracked)) == "" {
		return fmt.Errorf("file %q is not tracked by git", path)
	}

	var status bytes.Buffer
	cmd := exec.CommandContext(
		ctx, "git", "--literal-pathspecs", "-C", root, "status", "--porcelain", "--", relative,
	)
	cmd.Stdout = &status
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("checking git status for %q: %w", path, err)
	}
	if hasUnstagedChanges(status.String()) {
		return fmt.Errorf("%q has unstaged changes", path)
	}

	return nil
}

// hasUnstagedChanges reports whether the Git porcelain status contains a worktree change;
// staged-only changes do not count.
func hasUnstagedChanges(status string) bool {
	for line := range strings.SplitSeq(status, "\n") {
		if line == "" {
			continue
		}
		if len(line) < 2 || line[1] != ' ' {
			return true
		}
	}
	return false
}
