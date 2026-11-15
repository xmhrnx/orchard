package core

import (
	"testing"

	"orchard/internal/store"
)

func repo(name, origin, group string) *RepoState {
	return &RepoState{Repo: store.Repo{ID: name, Name: name, Group: group}, Origin: origin}
}

func TestSectionsByHost(t *testing.T) {
	repos := []*RepoState{
		repo("a", "git@gitlab.com:me/a.git", ""),
		repo("b", "https://github.com/me/b.git", ""),
		repo("c", "", ""),
		repo("d", "/srv/git/d.git", ""),
		repo("e", "https://github.com/me/e", ""),
	}
	secs := Sections(repos, GroupByHost)
	var titles []string
	for _, s := range secs {
		titles = append(titles, s.Title)
	}
	want := []string{"github.com", "gitlab.com", "Local and other remotes", "No remote"}
	if len(titles) != len(want) {
		t.Fatalf("sections = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Errorf("section %d = %q, want %q", i, titles[i], want[i])
		}
	}
	if len(secs[0].Repos) != 2 || secs[0].Repos[0].Name != "b" || secs[0].Repos[1].Name != "e" {
		t.Errorf("github.com repos out of order: %v", secs[0].Repos)
	}
}

func TestSectionsByGroupAndNone(t *testing.T) {
	repos := []*RepoState{repo("a", "", "Work"), repo("b", "", ""), repo("c", "", "Clients")}
	secs := Sections(repos, GroupByGroup)
	if len(secs) != 3 || secs[0].Title != "Clients" || secs[2].Title != "Ungrouped" {
		t.Fatalf("group sections = %+v", secs)
	}
	if flat := Sections(repos, GroupByNone); len(flat) != 1 || len(flat[0].Repos) != 3 {
		t.Fatalf("flat = %+v", flat)
	}
}
