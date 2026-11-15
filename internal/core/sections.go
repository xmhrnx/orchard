package core

import (
	"sort"
	"strings"

	"orchard/internal/forge"
)

// GroupBy is how the repository list is divided into sections.
type GroupBy string

// Ways to section the repository list.
const (
	GroupByHost  GroupBy = "host"
	GroupByGroup GroupBy = "group"
	GroupByNone  GroupBy = "none"
)

// Kinds of section, for choosing an icon.
const (
	SectionHost     = "host"
	SectionNoRemote = "no-remote"
	SectionLocal    = "local"
	SectionGroup    = "group"
	SectionAll      = "all"
)

// RepoSection is a titled run of repositories in the list.
type RepoSection struct {
	Key   string
	Kind  string
	Title string
	Repos []*RepoState
}

// Sections divides repositories, keeping their order within each
// section. Hosts and groups are sorted by name; repositories without a
// remote, with a local or unrecognised remote, or in no group come last.
func Sections(repos []*RepoState, by GroupBy) []RepoSection {
	if by != GroupByHost && by != GroupByGroup {
		return []RepoSection{{Key: SectionAll, Kind: SectionAll, Repos: repos}}
	}
	byKey := map[string]*RepoSection{}
	var order []string
	for _, r := range repos {
		sec := sectionOf(r, by)
		if existing, ok := byKey[sec.Key]; ok {
			existing.Repos = append(existing.Repos, r)
			continue
		}
		sec.Repos = []*RepoState{r}
		byKey[sec.Key] = &sec
		order = append(order, sec.Key)
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := byKey[order[i]], byKey[order[j]]
		if rank(a.Kind) != rank(b.Kind) {
			return rank(a.Kind) < rank(b.Kind)
		}
		return strings.ToLower(a.Title) < strings.ToLower(b.Title)
	})
	out := make([]RepoSection, len(order))
	for i, k := range order {
		out[i] = *byKey[k]
	}
	return out
}

// HostOf is where a repository's origin lives, or "" when it has no
// hosted origin.
func HostOf(r *RepoState) string {
	if loc, ok := forge.ParseRemote(r.Origin); ok {
		return loc.Host
	}
	return ""
}

func sectionOf(r *RepoState, by GroupBy) RepoSection {
	if by == GroupByGroup {
		if r.Group == "" {
			return RepoSection{Key: "group:", Kind: SectionNoRemote, Title: "Ungrouped"}
		}
		return RepoSection{Key: "group:" + r.Group, Kind: SectionGroup, Title: r.Group}
	}
	switch host := HostOf(r); {
	case host != "":
		return RepoSection{Key: "host:" + host, Kind: SectionHost, Title: host}
	case r.Origin == "":
		return RepoSection{Key: "host:", Kind: SectionNoRemote, Title: "No remote"}
	default:
		return RepoSection{Key: "host:local", Kind: SectionLocal, Title: "Local and other remotes"}
	}
}

// rank orders section kinds: named sections first, catch-alls last.
func rank(kind string) int {
	switch kind {
	case SectionLocal:
		return 1
	case SectionNoRemote:
		return 2
	}
	return 0
}
