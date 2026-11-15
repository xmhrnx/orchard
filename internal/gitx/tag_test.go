package gitx

import "testing"

func TestParseRemoteTagsKeepsObjectAndCommit(t *testing.T) {
	out := "aaa\trefs/tags/v1\nbbb\trefs/tags/v2\nccc\trefs/tags/v2^{}\nddd\trefs/heads/main\n"
	tags := parseRemoteTags(out)
	want := []RemoteTag{{"v1", "aaa", "aaa"}, {"v2", "bbb", "ccc"}}
	if len(tags) != 2 || tags[0] != want[0] || tags[1] != want[1] {
		t.Fatalf("tags = %+v, want %+v", tags, want)
	}
}

func TestPlanTagDeletions(t *testing.T) {
	local := []Tag{{Name: "same", Hash: "c1"}, {Name: "clash", Hash: "mine"}, {Name: "here", Hash: "c3"}}
	remote := []RemoteTag{{"same", "o1", "c1"}, {"clash", "o2", "theirs"}, {"there", "o4", "c4"}}
	names := []string{"same", "clash", "here", "there", "nowhere"}
	common := map[string]TagDeletion{
		"same":  {Name: "same", Local: true, RemoteObject: "o1"},
		"here":  {Name: "here", Local: true},
		"there": {Name: "there", RemoteObject: "o4"},
	}
	for conflict, clash := range map[TagConflict]TagDeletion{
		KeepOrigin: {Name: "clash", Local: true},
		KeepLocal:  {Name: "clash", RemoteObject: "o2"},
	} {
		want := map[string]TagDeletion{"clash": clash}
		for k, d := range common {
			want[k] = d
		}
		checkPlan(t, PlanTagDeletions(local, remote, names, conflict), want)
	}
}

func checkPlan(t *testing.T, plan []TagDeletion, want map[string]TagDeletion) {
	t.Helper()
	if len(plan) != len(want) {
		t.Fatalf("plan = %+v", plan)
	}
	for _, d := range plan {
		if d != want[d.Name] {
			t.Errorf("%s: %+v, want %+v", d.Name, d, want[d.Name])
		}
	}
}
