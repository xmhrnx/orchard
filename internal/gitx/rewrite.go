package gitx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNeedsFilterRepo is returned for rewrites only git-filter-repo can do.
var ErrNeedsFilterRepo = errors.New("this rewrite needs git-filter-repo (https://github.com/newren/git-filter-repo)")

// HasFilterRepo reports whether git-filter-repo is installed. It is far
// faster and safer than filter-branch, so rewrites prefer it.
func HasFilterRepo() bool {
	_, err := exec.LookPath("git-filter-repo")
	return err == nil
}

// RemovePathsFromHistory deletes files or folders from every commit on
// every branch and tag, as if they had never been committed.
func RemovePathsFromHistory(ctx context.Context, dir string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	if HasFilterRepo() {
		args := []string{"--invert-paths"}
		for _, p := range paths {
			args = append(args, "--path", p)
		}
		return filterRepo(ctx, dir, args...)
	}
	quoted := make([]string, len(paths))
	for i, p := range paths {
		quoted[i] = shellQuote(p)
	}
	filter := "git rm -r --cached --ignore-unmatch --quiet -- " + strings.Join(quoted, " ")
	return filterBranch(ctx, dir, "--index-filter", filter, "--prune-empty")
}

// StripLargeBlobs removes every file larger than size (e.g. "10M") from
// the whole history.
func StripLargeBlobs(ctx context.Context, dir, size string) error {
	if !HasFilterRepo() {
		return ErrNeedsFilterRepo
	}
	return filterRepo(ctx, dir, "--strip-blobs-bigger-than", size)
}

// ReplaceText rewrites every occurrence of each secret in every file of
// every commit with "***REMOVED***": for leaked passwords and keys.
func ReplaceText(ctx context.Context, dir string, secrets []string) error {
	if !HasFilterRepo() {
		return ErrNeedsFilterRepo
	}
	f, err := os.CreateTemp("", "orchard-replace-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	for _, s := range secrets {
		if _, err := fmt.Fprintln(f, s); err != nil {
			f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return filterRepo(ctx, dir, "--replace-text", f.Name())
}

// RewriteAuthor changes the name and email on every commit authored or
// committed with oldEmail.
func RewriteAuthor(ctx context.Context, dir, oldEmail, newName, newEmail string) error {
	if HasFilterRepo() {
		f, err := os.CreateTemp("", "orchard-mailmap-*")
		if err != nil {
			return err
		}
		defer os.Remove(f.Name())
		_, werr := fmt.Fprintf(f, "%s <%s> <%s>\n", newName, newEmail, oldEmail)
		if cerr := f.Close(); werr != nil || cerr != nil {
			return errors.Join(werr, cerr)
		}
		return filterRepo(ctx, dir, "--mailmap", f.Name())
	}
	script := fmt.Sprintf(`
if [ "$GIT_AUTHOR_EMAIL" = %[1]s ]; then export GIT_AUTHOR_NAME=%[2]s GIT_AUTHOR_EMAIL=%[3]s; fi
if [ "$GIT_COMMITTER_EMAIL" = %[1]s ]; then export GIT_COMMITTER_NAME=%[2]s GIT_COMMITTER_EMAIL=%[3]s; fi`,
		shellQuote(oldEmail), shellQuote(newName), shellQuote(newEmail))
	return filterBranch(ctx, dir, "--env-filter", script)
}

// SquashLast squashes the last n commits of the current branch into one
// with the given message.
func SquashLast(ctx context.Context, dir string, n int, message string) error {
	if n < 2 {
		return errors.New("squash needs at least two commits")
	}
	if _, err := run(ctx, dir, "reset", "--soft", fmt.Sprintf("HEAD~%d", n)); err != nil {
		return err
	}
	return CreateCommit(ctx, dir, CommitOptions{Message: message})
}

// RewordHead changes the message of the latest commit.
func RewordHead(ctx context.Context, dir, message string) error {
	_, err := runInput(ctx, dir, message, "commit", "--amend", "--only", "-F", "-")
	return err
}

// ForcePushAll replaces every branch and tag on remote with the local
// ones: the step that publishes a rewritten history.
func ForcePushAll(ctx context.Context, dir, remote string) error {
	if _, err := run(ctx, dir, "push", "--force", "--all", remote); err != nil {
		return err
	}
	_, err := run(ctx, dir, "push", "--force", "--tags", remote)
	return err
}

// filterRepo runs git-filter-repo and restores the remotes it removes by
// design, so the repository can still be pushed afterwards.
func filterRepo(ctx context.Context, dir string, args ...string) error {
	remotes, err := Remotes(ctx, dir)
	if err != nil {
		return err
	}
	if _, err := run(ctx, dir, append([]string{"filter-repo", "--force"}, args...)...); err != nil {
		return err
	}
	for _, r := range remotes {
		if _, err := run(ctx, dir, "remote", "add", r.Name, r.FetchURL); err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

// filterBranch is the fallback when git-filter-repo is missing. It
// rewrites all refs, then drops the backups and unreachable objects so
// the removed data is really gone from the local repository.
func filterBranch(ctx context.Context, dir string, args ...string) error {
	full := append([]string{"filter-branch", "--force"}, args...)
	full = append(full, "--tag-name-filter", "cat", "--", "--all")
	if _, err := run(ctx, dir, full...); err != nil {
		return err
	}
	return purgeBackups(ctx, dir)
}

func purgeBackups(ctx context.Context, dir string) error {
	out, err := run(ctx, dir, "for-each-ref", "--format=%(refname)", "refs/original/")
	if err != nil {
		return err
	}
	for _, ref := range strings.Fields(out) {
		if _, err := run(ctx, dir, "update-ref", "-d", ref); err != nil {
			return err
		}
	}
	if _, err := run(ctx, dir, "reflog", "expire", "--expire=now", "--all"); err != nil {
		return err
	}
	_, err = run(ctx, dir, "gc", "--prune=now", "--quiet")
	return err
}

// LargestFiles lists the biggest blobs ever committed, to find what to
// purge from history.
func LargestFiles(ctx context.Context, dir string, limit int) ([]BlobInfo, error) {
	objects, err := run(ctx, dir, "rev-list", "--objects", "--all")
	if err != nil {
		return nil, err
	}
	out, err := runInput(ctx, dir, objects, "cat-file", "--batch-check=%(objecttype) %(objectsize) %(rest)")
	if err != nil {
		return nil, err
	}
	return topBlobs(out, limit), nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(filepath.ToSlash(s), "'", `'\''`) + "'"
}
