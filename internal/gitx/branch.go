package gitx

import (
	"context"
	"strings"
)

// Branch is a local or remote-tracking branch.
type Branch struct {
	Name     string
	Remote   bool
	Current  bool
	Upstream string
	Track    string
	Hash     string
	Subject  string
}

// Branches lists local and remote-tracking branches.
func Branches(ctx context.Context, dir string) ([]Branch, error) {
	format := strings.Join([]string{"%(refname)", "%(HEAD)", "%(upstream:short)", "%(upstream:track)", "%(objectname:short)", "%(contents:subject)"}, fieldSep)
	out, err := run(ctx, dir, "for-each-ref", "--sort=-committerdate", "--format="+format, "refs/heads", "refs/remotes")
	if err != nil {
		return nil, err
	}
	var list []Branch
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Split(line, fieldSep)
		if len(f) < 6 || strings.HasSuffix(f[0], "/HEAD") {
			continue
		}
		b := Branch{Current: f[1] == "*", Upstream: f[2], Track: f[3], Hash: f[4], Subject: f[5]}
		if name, ok := strings.CutPrefix(f[0], "refs/heads/"); ok {
			b.Name = name
		} else {
			b.Name, b.Remote = strings.TrimPrefix(f[0], "refs/remotes/"), true
		}
		list = append(list, b)
	}
	return list, nil
}

// CurrentBranch returns the checked-out branch, or "" when detached.
func CurrentBranch(ctx context.Context, dir string) (string, error) {
	out, err := run(ctx, dir, "branch", "--show-current")
	return strings.TrimSpace(out), err
}

// Checkout switches to a branch. A remote branch ("origin/feature") is
// checked out as a new local branch tracking it.
func Checkout(ctx context.Context, dir string, b Branch) error {
	if b.Remote {
		local := b.Name[strings.Index(b.Name, "/")+1:]
		_, err := run(ctx, dir, "switch", "--track", "-c", local, b.Name)
		if err != nil && strings.Contains(err.Error(), "already exists") {
			_, err = run(ctx, dir, "switch", local)
		}
		return err
	}
	_, err := run(ctx, dir, "switch", b.Name)
	return err
}

// CreateBranch creates a branch at start (HEAD when empty), optionally
// switching to it.
func CreateBranch(ctx context.Context, dir, name, start string, checkout bool) error {
	args := []string{"branch", name}
	if checkout {
		args = []string{"switch", "-c", name}
	}
	if start != "" {
		args = append(args, start)
	}
	_, err := run(ctx, dir, args...)
	return err
}

// DeleteBranches removes local branches, or remote branches from their
// remote when the branch is remote-tracking.
func DeleteBranches(ctx context.Context, dir string, branches []Branch, force bool) error {
	var local []string
	remote := map[string][]string{}
	for _, b := range branches {
		if b.Remote {
			name, ref, _ := strings.Cut(b.Name, "/")
			remote[name] = append(remote[name], ref)
		} else {
			local = append(local, b.Name)
		}
	}
	if len(local) > 0 {
		flag := "-d"
		if force {
			flag = "-D"
		}
		if _, err := run(ctx, dir, append([]string{"branch", flag}, local...)...); err != nil {
			return err
		}
	}
	for name, refs := range remote {
		if _, err := run(ctx, dir, append([]string{"push", name, "--delete"}, refs...)...); err != nil {
			return err
		}
	}
	return nil
}

// RenameBranch renames a local branch.
func RenameBranch(ctx context.Context, dir, from, to string) error {
	_, err := run(ctx, dir, "branch", "-m", from, to)
	return err
}

// Merge merges ref into the current branch.
func Merge(ctx context.Context, dir, ref string, noFF bool) error {
	args := []string{"merge", ref}
	if noFF {
		args = append(args, "--no-ff")
	}
	_, err := run(ctx, dir, args...)
	return err
}

// Rebase rebases the current branch onto ref.
func Rebase(ctx context.Context, dir, ref string) error {
	_, err := run(ctx, dir, "rebase", ref)
	return err
}

// AbortOperation abandons an in-progress merge, rebase or cherry-pick.
func AbortOperation(ctx context.Context, dir string) error {
	for _, args := range [][]string{{"merge", "--abort"}, {"rebase", "--abort"}, {"cherry-pick", "--abort"}} {
		if _, err := run(ctx, dir, args...); err == nil {
			return nil
		}
	}
	return &Error{Args: []string{"abort"}, Stderr: "no merge, rebase or cherry-pick in progress"}
}

// PruneMergedBranches deletes local branches already merged into the
// current branch, except the current one and the usual trunk names.
func PruneMergedBranches(ctx context.Context, dir string) ([]string, error) {
	out, err := run(ctx, dir, "branch", "--merged", "--format=%(refname:short)%(HEAD)")
	if err != nil {
		return nil, err
	}
	keep := map[string]bool{"main": true, "master": true, "develop": true, "trunk": true}
	var gone []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" || strings.HasSuffix(line, "*") || keep[line] {
			continue
		}
		gone = append(gone, line)
	}
	if len(gone) == 0 {
		return nil, nil
	}
	_, err = run(ctx, dir, append([]string{"branch", "-d"}, gone...)...)
	return gone, err
}
