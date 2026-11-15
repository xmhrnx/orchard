package env

// Page is a top-level destination.
type Page int

// Pages of the app.
const (
	PageRepos Page = iota
	PageRepo
	PageClone
	PageAccounts
	PageSettings
)

// Route is where the app is: a page and, for PageRepo, which repository.
type Route struct {
	Page   Page
	RepoID string
}

// Navigator holds the current route and a back stack.
type Navigator struct {
	current  Route
	history  []Route
	onChange func(Route)
}

// OnChange registers the function that reacts to navigation.
func (n *Navigator) OnChange(fn func(Route)) { n.onChange = fn }

// Current is the route being shown.
func (n *Navigator) Current() Route { return n.current }

// Go navigates to a route.
func (n *Navigator) Go(r Route) {
	if r == n.current {
		return
	}
	n.history = append(n.history, n.current)
	n.current = r
	if n.onChange != nil {
		n.onChange(r)
	}
}

// Open shows a top-level page.
func (n *Navigator) Open(p Page) { n.Go(Route{Page: p}) }

// OpenRepo shows a repository.
func (n *Navigator) OpenRepo(id string) { n.Go(Route{Page: PageRepo, RepoID: id}) }

// Back returns to the previous route, if any.
func (n *Navigator) Back() {
	if len(n.history) == 0 {
		return
	}
	prev := n.history[len(n.history)-1]
	n.history = n.history[:len(n.history)-1]
	n.current = prev
	if n.onChange != nil {
		n.onChange(prev)
	}
}
