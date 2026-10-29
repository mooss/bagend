// Package gitw provides git wrappers.
package gitw

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mooss/bagend/go/errs"
	"github.com/mooss/bagend/go/obj/osl"
	"github.com/mooss/bagend/go/pipe"
)

// Repo represents a Git repository.
type Repo struct {
	ctx  context.Context
	root string
}

// New returns the repository containing the current working directory.
// The Repo is scoped to the lifetime of the context.
func New(ctx context.Context) (Repo, error) {
	repo, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		if ctx.Err() != nil {
			return Repo{}, ctx.Err()
		}
		return Repo{}, errors.New("cannot execute outside a git repository")
	}

	return Repo{ctx: ctx, root: strings.TrimSpace(string(repo))}, nil
}

// Root returns the top-level directory of the repository.
func (r Repo) Root() string {
	return r.root
}

///////////////////////
// Public operations //

// FileIsClean requires the file to be tracked and inside the repository, and rejects files with
// unstaged changes.
func (r Repo) FileIsClean(path string) error {
	return pipe.Try1(r.relativeToRoot, path).
		Validate(r.ensureTracked).
		Validate(r.ensureNoUnstagedChanges).
		Err()
}

/////////////////////////
// Internal operations //

// relativeToRoot resolves path to a repository-relative path, rejecting paths outside the
// repository root.
func (r Repo) relativeToRoot(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %q: %w", path, err)
	}

	relative, err := filepath.Rel(r.root, absolute)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q is outside the git repository %q", path, r.root)
	}

	return relative, nil
}

// ensureTracked rejects files that are not tracked by git.
func (r Repo) ensureTracked(relative string) error {
	tracked, err := r.lsFiles(relative)
	if err != nil {
		return err
	}

	if strings.TrimSpace(tracked) == "" {
		return fmt.Errorf("file %q is not tracked by git", relative)
	}

	return nil
}

// ensureNoUnstagedChanges rejects files with unstaged changes.
func (r Repo) ensureNoUnstagedChanges(relative string) error {
	status, err := r.status(relative)
	if err != nil {
		return err
	}

	if hasUnstagedChanges(status) {
		return fmt.Errorf("%q has unstaged changes", relative)
	}

	return nil
}

/////////////////////
// Utility methods //

func (r Repo) run(args ...string) (string, error) {
	args = *slc("--literal-pathspecs", "-C", r.root).Cat(args)
	res, err := exec.CommandContext(r.ctx, "git", args...).Output()
	return string(res), err
}

func (r Repo) lsFiles(args ...string) (string, error) {
	wrap := errs.Wrapper("checking git tracking of %v", args)
	args = *slc("ls-files", "--").Cat(args)
	return wrap.Result(r.run(args...))
}

func (r Repo) status(args ...string) (string, error) {
	wrap := errs.Wrapper("checking git status of %v", args)
	args = *slc("status", "--porcelain", "--").Cat(args)
	return wrap.Result(r.run(args...))
}

///////////////////////
// Utility functions //

func slc(args ...string) *osl.Slice[string] { return osl.NewP(args...) }

// hasUnstagedChanges reports whether the porcelain status contains a worktree change.
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
