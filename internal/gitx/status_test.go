package gitx

import "testing"

func TestParseStatus(t *testing.T) {
	out := "# branch.oid abc123\x00# branch.head main\x00# branch.upstream origin/main\x00# branch.ab +2 -3\x00# stash 1\x00" +
		"1 M. N... 100644 100644 100644 aaa bbb staged.go\x00" +
		"1 .M N... 100644 100644 100644 aaa bbb edited.go\x00" +
		"2 R. N... 100644 100644 100644 aaa bbb R100 new.go\x00old.go\x00" +
		"? draft.txt\x00"
	st := parseStatus(out)
	if st.Branch != "main" || st.Upstream != "origin/main" || st.Ahead != 2 || st.Behind != 3 || st.Stashes != 1 {
		t.Fatalf("header parsed wrong: %+v", st)
	}
	if len(st.Staged) != 2 || st.Staged[1].Kind != Renamed || st.Staged[1].OrigPath != "old.go" {
		t.Errorf("staged = %+v", st.Staged)
	}
	if len(st.Unstaged) != 1 || st.Unstaged[0].Path != "edited.go" {
		t.Errorf("unstaged = %+v", st.Unstaged)
	}
	if len(st.Untracked) != 1 || st.ChangeCount() != 4 {
		t.Errorf("untracked = %+v, count %d", st.Untracked, st.ChangeCount())
	}
}
