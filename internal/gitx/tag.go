package gitx

import (
	"context"
	"strings"
	"time"
)

// Tag is a lightweight or annotated tag.
type Tag struct {
	Name    string
	Hash    string
	Subject string
	When    time.Time
}

// Tags lists tags, newest first.
func Tags(ctx context.Context, dir string) ([]Tag, error) {
	format := strings.Join([]string{"%(refname:short)", "%(objectname:short)", "%(contents:subject)", "%(creatordate:unix)"}, fieldSep)
	out, err := run(ctx, dir, "for-each-ref", "--sort=-creatordate", "--format="+format, "refs/tags")
	if err != nil {
		return nil, err
	}
	var tags []Tag
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Split(line, fieldSep)
		if len(f) < 4 {
			continue
		}
		tags = append(tags, Tag{Name: f[0], Hash: f[1], Subject: f[2], When: parseUnix(f[3])})
	}
	return tags, nil
}

// CreateTag tags ref (HEAD when empty); a message makes it annotated.
func CreateTag(ctx context.Context, dir, name, ref, message string) error {
	args := []string{"tag"}
	if message != "" {
		args = append(args, "-a", "-m", message)
	}
	args = append(args, name)
	if ref != "" {
		args = append(args, ref)
	}
	_, err := run(ctx, dir, args...)
	return err
}

// DeleteTags removes tags locally and, with remote set, from origin too,
// in a single push so hundreds of tags go in one round trip.
func DeleteTags(ctx context.Context, dir string, names []string, remote string) error {
	if len(names) == 0 {
		return nil
	}
	if remote != "" {
		refs := make([]string, len(names))
		for i, n := range names {
			refs[i] = ":refs/tags/" + n
		}
		if _, err := run(ctx, dir, append([]string{"push", remote}, refs...)...); err != nil {
			return err
		}
	}
	_, err := run(ctx, dir, append([]string{"tag", "-d"}, names...)...)
	return err
}

// DeleteTagsIfPresent deletes tags from origin and locally, skipping
// each place a tag doesn't exist, so cleaning up after releases never
// fails on a tag that was already gone.
func DeleteTagsIfPresent(ctx context.Context, dir string, names []string) error {
	local, err := Tags(ctx, dir)
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, t := range local {
		have[t.Name] = true
	}
	out, err := run(ctx, dir, "ls-remote", "--tags", "origin")
	if err != nil {
		return err
	}
	remote := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) == 2 {
			remote[strings.TrimSuffix(strings.TrimPrefix(f[1], "refs/tags/"), "^{}")] = true
		}
	}
	var refs, locals []string
	for _, n := range names {
		if remote[n] {
			refs = append(refs, ":refs/tags/"+n)
		}
		if have[n] {
			locals = append(locals, n)
		}
	}
	if len(refs) > 0 {
		if _, err := run(ctx, dir, append([]string{"push", "origin"}, refs...)...); err != nil {
			return err
		}
	}
	if len(locals) > 0 {
		_, err = run(ctx, dir, append([]string{"tag", "-d"}, locals...)...)
	}
	return err
}
