package graph

import "testing"

func TestLinearHistoryStaysInOneLane(t *testing.T) {
	l := Compute([]Node{{"c", []string{"b"}}, {"b", []string{"a"}}, {"a", nil}})
	if l.Lanes != 1 {
		t.Fatalf("lanes = %d, want 1", l.Lanes)
	}
	for i, r := range l.Rows {
		if r.Lane != 0 {
			t.Errorf("row %d in lane %d", i, r.Lane)
		}
	}
	if got := len(l.Rows[2].Segments); got != 1 {
		t.Errorf("root row has %d segments, want only the line from above", got)
	}
}

func TestMergeOpensAndClosesALane(t *testing.T) {
	// m merges f into b; both descend from a.
	l := Compute([]Node{
		{"m", []string{"b", "f"}},
		{"f", []string{"a"}},
		{"b", []string{"a"}},
		{"a", nil},
	})
	if l.Lanes != 2 {
		t.Fatalf("lanes = %d, want 2", l.Lanes)
	}
	if l.Rows[1].Lane != 1 {
		t.Errorf("feature commit in lane %d, want 1", l.Rows[1].Lane)
	}
	if l.Rows[3].Lane != 0 {
		t.Errorf("root in lane %d, want 0", l.Rows[3].Lane)
	}
	converging := 0
	for _, s := range l.Rows[3].Segments {
		if s.Half == Top {
			converging++
		}
	}
	if converging != 2 {
		t.Errorf("root has %d incoming lines, want 2", converging)
	}
}

func TestTwoHeadsGetSeparateLanes(t *testing.T) {
	l := Compute([]Node{{"x", []string{"a"}}, {"y", []string{"a"}}, {"a", nil}})
	if l.Rows[0].Lane == l.Rows[1].Lane {
		t.Errorf("independent heads share lane %d", l.Rows[0].Lane)
	}
	if l.Rows[0].Color == l.Rows[1].Color {
		t.Error("independent heads share a color")
	}
}

func TestMissingParentsCloseTheirLanes(t *testing.T) {
	// A shallow history: b's parent "gone" was never fetched, and c on
	// another branch also points outside the range.
	l := Compute([]Node{{"a", []string{"b"}}, {"b", []string{"gone"}}, {"c", []string{"gone2"}}, {"d", nil}})
	if l.Lanes != 1 {
		t.Fatalf("lanes = %d, want 1: lanes waiting for missing parents must close", l.Lanes)
	}
}
