// Package screens holds Orchard's pages and the views inside them, one
// component per file.
package screens

import (
	"strings"

	"orchard/internal/core"
)

// repoFilter is which repositories the overview shows.
type repoFilter int

const (
	filterAll repoFilter = iota
	filterChanged
	filterBehind
	filterAhead
	filterIssues
)

var repoFilterNames = []string{"All", "Changed", "Behind", "Ahead", "Issues"}

// matches reports whether a repository passes the filter and query.
func (f repoFilter) matches(r *core.RepoState, query string) bool {
	if q := strings.ToLower(strings.TrimSpace(query)); q != "" {
		hay := strings.ToLower(r.Name + " " + r.Path + " " + r.Status.Branch + " " + r.Group)
		if !strings.Contains(hay, q) {
			return false
		}
	}
	st := r.Status
	switch f {
	case filterChanged:
		return st.Dirty()
	case filterBehind:
		return st.Behind > 0
	case filterAhead:
		return st.Ahead > 0
	case filterIssues:
		return r.Err != nil || (r.Loaded && st.Upstream == "" && !st.Detached) || len(st.Conflicted) > 0
	}
	return true
}
