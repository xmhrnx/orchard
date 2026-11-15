package gitx

import "testing"

func TestSplitRefsKeepsSlashesInLocalBranches(t *testing.T) {
	refs := splitRefs("HEAD -> refs/heads/main, tag: refs/tags/v1.0, refs/heads/feature/lanes, refs/remotes/origin/main, refs/remotes/origin/HEAD")
	want := []Ref{{"main", RefBranch}, {"v1.0", RefTag}, {"feature/lanes", RefBranch}, {"origin/main", RefRemote}}
	if len(refs) != len(want) {
		t.Fatalf("got %v, want %v", refs, want)
	}
	for i := range want {
		if refs[i] != want[i] {
			t.Errorf("ref %d = %v, want %v", i, refs[i], want[i])
		}
	}
}
