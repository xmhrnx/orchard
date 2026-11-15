package screens

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/core"
)

// repoView is a tab of the repository page. Open (re)loads it for a
// repository; it is called when the page opens and when the tab is
// chosen.
type repoView interface {
	giode.Element
	Open(r *core.RepoState)
}
