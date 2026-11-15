package gitx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newRepo creates a repository with a few commits in a temp folder.
func newRepo(t *testing.T) string {
	t.Helper()
	if !Available() {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	sh := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=T", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=T", "GIT_COMMITTER_EMAIL=t@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sh("init", "-q", "-b", "main")
	write("app.go", "package app\n")
	write("secret.env", "TOKEN=hunter2\n")
	sh("add", ".")
	sh("commit", "-qm", "first")
	write("app.go", "package app\n\nfunc A() {}\n")
	sh("commit", "-qam", "second")
	for _, tag := range []string{"v1", "v2", "v3"} {
		sh("tag", tag)
	}
	return dir
}

func TestStatusAndLogOnRealRepository(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := ReadStatus(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if st.Branch != "main" || len(st.Untracked) != 1 {
		t.Errorf("status = %+v", st)
	}
	commits, err := Log(ctx, dir, LogOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 || commits[0].Subject != "second" {
		t.Fatalf("log = %+v", commits)
	}
	if len(commits[0].Refs) == 0 {
		t.Error("head commit has no refs")
	}
}

func TestDeleteTagsRemovesSeveralAtOnce(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()
	if err := DeleteTags(ctx, dir, []string{"v1", "v3"}, ""); err != nil {
		t.Fatal(err)
	}
	tags, err := Tags(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0].Name != "v2" {
		t.Fatalf("tags = %+v, want only v2", tags)
	}
}

func TestRemovePathsFromHistoryErasesTheFile(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()
	if err := RemovePathsFromHistory(ctx, dir, []string{"secret.env"}); err != nil {
		t.Fatal(err)
	}
	out, err := run(ctx, dir, "log", "--all", "--name-only", "--format=")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "secret.env") {
		t.Fatalf("secret.env still in history:\n%s", out)
	}
	if !strings.Contains(out, "app.go") {
		t.Fatal("rewrite lost unrelated files")
	}
}

func TestSquashLastCombinesCommits(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()
	if _, err := run(ctx, dir, "-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "--allow-empty", "-qm", "third"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_AUTHOR_NAME", "T")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "T")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")
	if err := SquashLast(ctx, dir, 2, "combined"); err != nil {
		t.Fatal(err)
	}
	commits, err := Log(ctx, dir, LogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 || commits[0].Subject != "combined" {
		t.Fatalf("log after squash = %+v", commits)
	}
}

func TestPullWithoutUpstreamSaysSo(t *testing.T) {
	dir := newRepo(t)
	if err := Pull(context.Background(), dir, PullFastForward, true); err != ErrNoUpstream {
		t.Fatalf("Pull = %v, want ErrNoUpstream", err)
	}
}

func TestDeleteTagsIfPresentSkipsMissingTags(t *testing.T) {
	origin := newRepo(t)
	ctx := context.Background()
	clone := filepath.Join(t.TempDir(), "clone")
	if err := Clone(ctx, CloneOptions{URL: origin, Dest: clone}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := run(ctx, origin, "tag", "remote-only"); err != nil {
		t.Fatal(err)
	}
	if err := DeleteTagsIfPresent(ctx, clone, []string{"v1", "remote-only", "nowhere"}); err != nil {
		t.Fatal(err)
	}
	for dir, gone := range map[string][]string{origin: {"v1", "remote-only"}, clone: {"v1"}} {
		tags, err := Tags(ctx, dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, tg := range tags {
			for _, g := range gone {
				if tg.Name == g {
					t.Errorf("%s still has tag %s", dir, g)
				}
			}
		}
	}
}

func TestRemoteTagsAndFetchTagsBringInTagsAddedUpstream(t *testing.T) {
	origin := newRepo(t)
	ctx := context.Background()
	clone := filepath.Join(t.TempDir(), "clone")
	if err := Clone(ctx, CloneOptions{URL: origin, Dest: clone}, nil); err != nil {
		t.Fatal(err)
	}
	// A tag made upstream on a commit that is on no branch (its branch
	// was deleted): a plain fetch doesn't follow it.
	for _, args := range [][]string{
		{"switch", "-q", "-c", "gone"},
		{"-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "on a branch that goes away"},
		{"-c", "user.name=T", "-c", "user.email=t@example.com", "tag", "-a", "-m", "late", "late"},
		{"switch", "-q", "main"},
		{"branch", "-q", "-D", "gone"},
	} {
		if _, err := run(ctx, origin, args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := run(ctx, clone, "fetch", "origin"); err != nil {
		t.Fatal(err)
	}
	if hasTag(t, clone, "late") {
		t.Fatal("expected a plain fetch to leave the tag behind")
	}
	remote, err := RemoteTags(ctx, clone, "origin")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range remote {
		found = found || r.Name == "late"
	}
	if !found {
		t.Fatalf("remote tags %v lack the new tag", remote)
	}
	if err := FetchTags(ctx, clone, "origin"); err != nil {
		t.Fatal(err)
	}
	if !hasTag(t, clone, "late") {
		t.Fatal("FetchTags didn't bring the tag in")
	}
}

func hasTag(t *testing.T, dir, name string) bool {
	t.Helper()
	tags, err := Tags(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tg := range tags {
		if tg.Name == name {
			return true
		}
	}
	return false
}

func TestCreateTagIsAnnotatedAndDatedNow(t *testing.T) {
	dir := newRepo(t)
	ctx := context.Background()
	t.Setenv("GIT_COMMITTER_NAME", "T")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")
	if err := CreateTag(ctx, dir, "fresh", "HEAD~1", ""); err != nil {
		t.Fatal(err)
	}
	tags, err := Tags(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if tags[0].Name != "fresh" || !tags[0].Annotated || tags[0].Subject != "fresh" || time.Since(tags[0].When) > time.Minute {
		t.Fatalf("newest tag = %+v; want the new annotated tag, dated now", tags[0])
	}
	head1, _ := run(ctx, dir, "rev-parse", "--short", "HEAD~1")
	if !strings.HasPrefix(strings.TrimSpace(head1), tags[0].Hash) && !strings.HasPrefix(tags[0].Hash, strings.TrimSpace(head1)) {
		t.Errorf("tag hash %s is not the tagged commit %s", tags[0].Hash, head1)
	}
}

func TestPushTagPublishesOneTag(t *testing.T) {
	origin := newRepo(t)
	ctx := context.Background()
	clone := filepath.Join(t.TempDir(), "clone")
	if err := Clone(ctx, CloneOptions{URL: origin, Dest: clone}, nil); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_COMMITTER_NAME", "T")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")
	if err := CreateTag(ctx, clone, "shipped", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := PushTag(ctx, clone, "origin", "shipped"); err != nil {
		t.Fatal(err)
	}
	if !hasTag(t, origin, "shipped") {
		t.Fatal("tag not on origin after PushTag")
	}
}

// TestDeletingASameNamedLocalTagKeepsOrigins reproduces deleting, from
// the Tags tab, a local tag that shares its name with a different tag on
// origin: origin's tag must survive.
func TestDeletingASameNamedLocalTagKeepsOrigins(t *testing.T) {
	origin := newRepo(t)
	ctx := context.Background()
	t.Setenv("GIT_COMMITTER_NAME", "T")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")
	clone := filepath.Join(t.TempDir(), "clone")
	if err := Clone(ctx, CloneOptions{URL: origin, Dest: clone}, nil); err != nil {
		t.Fatal(err)
	}
	if err := CreateTag(ctx, origin, "v9", "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if err := CreateTag(ctx, clone, "v9", "HEAD~1", ""); err != nil {
		t.Fatal(err)
	}
	local, _ := Tags(ctx, clone)
	remote, err := RemoteTags(ctx, clone, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteTagsSafely(ctx, clone, "origin", PlanTagDeletions(local, remote, []string{"v9"}, KeepOrigin)); err != nil {
		t.Fatal(err)
	}
	if !hasTag(t, origin, "v9") {
		t.Fatal("origin's different v9 was deleted")
	}
	if hasTag(t, clone, "v9") {
		t.Fatal("the local v9 survived")
	}
}

func TestRemoteDeleteIsRefusedWhenOriginsTagChanged(t *testing.T) {
	origin := newRepo(t)
	ctx := context.Background()
	clone := filepath.Join(t.TempDir(), "clone")
	if err := Clone(ctx, CloneOptions{URL: origin, Dest: clone}, nil); err != nil {
		t.Fatal(err)
	}
	stale, err := RemoteTags(ctx, clone, "origin")
	if err != nil {
		t.Fatal(err)
	}
	// Someone moves v1 on origin after Orchard looked.
	if _, err := run(ctx, origin, "tag", "-f", "v1", "HEAD~1"); err != nil {
		t.Fatal(err)
	}
	var obj string
	for _, r := range stale {
		if r.Name == "v1" {
			obj = r.Object
		}
	}
	err = DeleteTagsSafely(ctx, clone, "origin", []TagDeletion{{Name: "v1", RemoteObject: obj}})
	if err == nil {
		t.Fatal("delete went through although origin's tag changed")
	}
	if !hasTag(t, origin, "v1") {
		t.Fatal("origin's moved tag was deleted")
	}
}
