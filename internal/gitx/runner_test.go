package gitx

import "testing"

func TestLastLinePrefersFatal(t *testing.T) {
	out := "hint: something\nfatal: not possible to fast-forward, aborting.\nhint: more\n"
	if got := lastLine(out); got != "not possible to fast-forward, aborting." {
		t.Errorf("lastLine = %q", got)
	}
	if got := lastLine("one\ntwo\nhint: x"); got != "two" {
		t.Errorf("lastLine = %q, want the last non-hint line", got)
	}
}
