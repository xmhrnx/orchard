package screens

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/segmented"

	"orchard/ui/components"
	"orchard/ui/env"
)

// CloneScreen brings repositories onto this computer: one by URL, or
// many at once from a connected account.
type CloneScreen struct {
	env     *env.Env
	mode    *segmented.Control
	form    *CloneForm
	browser *RemoteBrowser
	jobs    *CloneJobs
	scroll  *scroll.Scroll
}

// NewCloneScreen creates the page. Create it once.
func NewCloneScreen(e *env.Env) *CloneScreen {
	s := &CloneScreen{env: e, jobs: NewCloneJobs(e)}
	s.form = NewCloneForm(e, s.jobs)
	s.browser = NewRemoteBrowser(e, s.jobs)
	s.mode = giode.Input.Segmented("From a URL", "From an account").OnChange(func(i int) {
		if i == 1 {
			s.browser.Prefill()
		}
	})
	s.mode.Icons(components.Glyph("link", 14, ""), components.Glyph("user", 14, ""))
	s.scroll = giode.Scroll(func() giode.Element {
		return giode.Stack(giode.Styles{Width: -1, Gap: 18}, s.form, s.jobs)
	}, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	return s
}

// Layout draws the page.
func (s *CloneScreen) Layout(gtx giode.Context) giode.Dimensions {
	s.form.Prefill()
	body := giode.Element(giode.Box(giode.Styles{Width: -1, FlexGrow: 1, MaxWidth: 920}, s.scroll))
	if s.mode.Selected() == 1 {
		body = giode.Stack(giode.Styles{Width: -1, FlexGrow: 1, Gap: 14},
			giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, s.browser),
			s.jobs)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 20, Padding: giode.Inset{Top: 26, Left: 32, Right: 32, Bottom: 20}},
		components.PageHeader("Clone", "Plant new trees. Anything you clone joins your orchard automatically.", s.mode),
		body,
	).Layout(gtx)
}
