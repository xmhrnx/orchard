package forge

import "testing"

func TestParseRemote(t *testing.T) {
	cases := map[string]RemoteLocation{
		"https://github.com/owner/name.git":         {"github.com", "owner/name"},
		"git@gitlab.com:group/sub/name.git":         {"gitlab.com", "group/sub/name"},
		"ssh://git@codeberg.org:22/owner/name":      {"codeberg.org", "owner/name"},
		"https://user@bitbucket.org/team/repo.git/": {"bitbucket.org", "team/repo"},
	}
	for in, want := range cases {
		got, ok := ParseRemote(in)
		if !ok || got != want {
			t.Errorf("ParseRemote(%q) = %+v, %v; want %+v", in, got, ok, want)
		}
	}
	if _, ok := ParseRemote("/srv/git/repo"); ok {
		t.Error("a local path parsed as a hosted remote")
	}
}
