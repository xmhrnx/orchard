package gitx

import (
	"context"
	"strings"
	"time"
)

// Tag is a lightweight or annotated tag. An annotated tag has its own
// date and message; a lightweight one is only a name, so When and
// Subject are those of the commit it points at.
type Tag struct {
	Name string
	// Hash is the full hash of the commit the tag points at.
	Hash      string
	Subject   string
	When      time.Time
	Annotated bool
}

// RemoteTag is a tag as a remote has it: Object is what the tag ref
// holds (the tag object for annotated tags), Commit what it points at.
type RemoteTag struct {
	Name   string
	Object string
	Commit string
}

// TagConflict chooses which copy "delete everywhere" removes when a tag
// with the same name points at different commits here and on origin.
type TagConflict int

const (
	// KeepOrigin removes the local tag and leaves origin's: deleting
	// from the Tags tab never destroys a different tag on the server.
	KeepOrigin TagConflict = iota
	// KeepLocal removes origin's tag and leaves the local one: for
	// cleaning up after a release, whose tag is origin's.
	KeepLocal
)

// TagDeletion is one tag to delete. Local removes it here; a non-empty
// RemoteObject removes it from the remote too, but only while the
// remote's tag is still exactly that object, so a different tag that
// happens to share the name is never deleted.
type TagDeletion struct {
	Name         string
	Local        bool
	RemoteObject string
}

// Tags lists tags, newest first.
func Tags(ctx context.Context, dir string) ([]Tag, error) {
	format := strings.Join([]string{"%(refname:short)", "%(objectname)", "%(contents:subject)", "%(creatordate:unix)", "%(objecttype)", "%(*objectname)"}, fieldSep)
	out, err := run(ctx, dir, "for-each-ref", "--sort=-creatordate", "--format="+format, "refs/tags")
	if err != nil {
		return nil, err
	}
	var tags []Tag
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Split(line, fieldSep)
		if len(f) < 6 {
			continue
		}
		hash := f[1]
		if f[5] != "" {
			hash = f[5] // an annotated tag: show the commit it points at
		}
		tags = append(tags, Tag{Name: f[0], Hash: hash, Subject: f[2], When: parseUnix(f[3]), Annotated: f[4] == "tag"})
	}
	return tags, nil
}

// CreateTag makes an annotated tag on ref (HEAD when empty). Annotated
// tags record when and by whom they were made, so a new tag never
// borrows an old commit's date; an empty message uses the tag's name.
func CreateTag(ctx context.Context, dir, name, ref, message string) error {
	if message == "" {
		message = name
	}
	args := []string{"tag", "-a", "-m", message, name}
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

// DeleteTagsSafely deletes tags here and on a remote in one push. Each
// remote deletion is leased on the object the caller saw there, so the
// server refuses it if that tag was replaced in the meantime.
func DeleteTagsSafely(ctx context.Context, dir, remote string, dels []TagDeletion) error {
	var leases, refs, local []string
	for _, d := range dels {
		if d.RemoteObject != "" {
			leases = append(leases, "--force-with-lease=refs/tags/"+d.Name+":"+d.RemoteObject)
			refs = append(refs, ":refs/tags/"+d.Name)
		}
		if d.Local {
			local = append(local, d.Name)
		}
	}
	if len(refs) > 0 {
		args := append(append([]string{"push"}, leases...), remote)
		if _, err := run(ctx, dir, append(args, refs...)...); err != nil {
			return err
		}
	}
	if len(local) > 0 {
		_, err := run(ctx, dir, append([]string{"tag", "-d"}, local...)...)
		return err
	}
	return nil
}

// DeleteTagsIfPresent deletes released tags from origin and here,
// wherever they exist. A local tag that differs from origin's is kept.
func DeleteTagsIfPresent(ctx context.Context, dir string, names []string) error {
	local, err := Tags(ctx, dir)
	if err != nil {
		return err
	}
	remote, err := RemoteTags(ctx, dir, "origin")
	if err != nil {
		return err
	}
	return DeleteTagsSafely(ctx, dir, "origin", PlanTagDeletions(local, remote, names, KeepLocal))
}

// PlanTagDeletions decides what deleting each named tag "everywhere"
// removes. A tag that is the same here and on origin (same commit) goes
// from both; a tag on one side only goes from that side; a tag that
// differs between the two loses only the copy conflict doesn't keep.
func PlanTagDeletions(local []Tag, remote []RemoteTag, names []string, conflict TagConflict) []TagDeletion {
	here := map[string]Tag{}
	for _, t := range local {
		here[t.Name] = t
	}
	there := map[string]RemoteTag{}
	for _, r := range remote {
		there[r.Name] = r
	}
	var plan []TagDeletion
	for _, n := range names {
		l, isLocal := here[n]
		r, isRemote := there[n]
		d := TagDeletion{Name: n, Local: isLocal}
		if isRemote {
			d.RemoteObject = r.Object
		}
		if isLocal && isRemote && l.Hash != r.Commit {
			if conflict == KeepOrigin {
				d.RemoteObject = ""
			} else {
				d.Local = false
			}
		}
		if d.Local || d.RemoteObject != "" {
			plan = append(plan, d)
		}
	}
	return plan
}

// RemoteTags lists the tags on a remote. Annotated tags report the
// commit they point at.
func RemoteTags(ctx context.Context, dir, remote string) ([]RemoteTag, error) {
	out, err := run(ctx, dir, "ls-remote", "--tags", remote)
	if err != nil {
		return nil, err
	}
	return parseRemoteTags(out), nil
}

func parseRemoteTags(out string) []RemoteTag {
	byName := map[string]int{}
	var tags []RemoteTag
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Fields(line)
		if len(f) != 2 || !strings.HasPrefix(f[1], "refs/tags/") {
			continue
		}
		name, peeled := strings.CutSuffix(strings.TrimPrefix(f[1], "refs/tags/"), "^{}")
		i, seen := byName[name]
		if !seen {
			i = len(tags)
			byName[name] = i
			tags = append(tags, RemoteTag{Name: name})
		}
		if peeled {
			tags[i].Commit = f[0]
			continue
		}
		tags[i].Object = f[0]
		if tags[i].Commit == "" {
			tags[i].Commit = f[0]
		}
	}
	return tags
}

// FetchTags downloads every tag from a remote, including tags on commits
// already present, which a plain fetch leaves behind.
func FetchTags(ctx context.Context, dir, remote string) error {
	_, err := run(ctx, dir, "fetch", "--tags", remote)
	return err
}

// PushTag sends one tag to a remote.
func PushTag(ctx context.Context, dir, remote, name string) error {
	_, err := run(ctx, dir, "push", remote, "refs/tags/"+name)
	return err
}
