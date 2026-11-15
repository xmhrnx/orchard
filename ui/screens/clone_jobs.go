package screens

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// cloneJob is one clone in flight or finished.
type cloneJob struct {
	name     string
	dest     string
	progress gitx.Progress
	done     bool
	err      error
}

// CloneJobs runs clones and shows their live progress.
type CloneJobs struct {
	env  *env.Env
	jobs []*cloneJob
}

// NewCloneJobs creates the list. Create it once.
func NewCloneJobs(e *env.Env) *CloneJobs { return &CloneJobs{env: e} }

// Start clones a repository, showing its progress in the list.
func (c *CloneJobs) Start(req core.CloneRequest) {
	job := &cloneJob{name: gitx.RepoNameFromURL(req.URL), dest: req.Dest(), progress: gitx.Progress{Phase: "Starting"}}
	c.jobs = append([]*cloneJob{job}, c.jobs...)
	c.env.WS.Clone(req, func(p gitx.Progress) { job.progress = p }, func(err error) {
		job.done, job.err = true, err
		if err != nil {
			c.env.Fail("Cloning "+job.name, err)
		} else {
			c.env.Toasts.Show(giode.ToastSuccess, job.name+" is ready")
		}
	})
}

// Active reports how many clones are still running.
func (c *CloneJobs) Active() int {
	n := 0
	for _, j := range c.jobs {
		if !j.done {
			n++
		}
	}
	return n
}

func (c *CloneJobs) row(j *cloneJob) giode.Element {
	status := giode.Element(giode.Text(fmt.Sprintf("%s · %d%%", j.progress.Phase, j.progress.Percent), theme.Caption))
	icon, tint := "clone", theme.Primary
	switch {
	case j.err != nil:
		icon, tint = "alert", theme.Danger
		status = giode.Text(j.err.Error(), theme.Ellipsis(theme.Tinted(theme.Caption, theme.Danger)))
	case j.done:
		icon = "check"
		status = giode.Text("Cloned to "+components.ShortPath(j.dest), theme.Ellipsis(theme.Caption))
	}
	kids := []giode.Element{
		giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			components.Glyph(icon, 16, tint),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(j.name, theme.Ellipsis(theme.Strong)))),
		status,
	}
	if !j.done {
		kids = append(kids, giode.Progress(float32(j.progress.Percent)/100, giode.Styles{Color: theme.Primary, Background: theme.Sunken, Height: 4}))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6, Padding: giode.UniformInset(12), BorderRadius: theme.Radius, Background: theme.Raised}, kids...)
}

// Layout draws the list; it is empty until a clone starts.
func (c *CloneJobs) Layout(gtx giode.Context) giode.Dimensions {
	if len(c.jobs) == 0 {
		return giode.Box(giode.Styles{}).Layout(gtx)
	}
	rows := []giode.Element{giode.Text("Clones", theme.Overline)}
	for _, j := range c.jobs {
		rows = append(rows, c.row(j))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, rows...).Layout(gtx)
}
