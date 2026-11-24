package screens

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"orchard/internal/project"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// BuildTargetsPanel is the build dropdown: every giode target, checked
// against giode doctor, with the ones to build ticked. The choice is
// saved per repository at once, so Build needs no setup next time.
type BuildTargetsPanel struct {
	env     *env.Env
	repoID  func() string
	doctor  *doctorCache
	boxes   map[string]*checkbox.Checkbox
	recheck *components.IconAction
	build   *giode.ButtonElement
}

// NewBuildTargetsPanel creates the panel; build starts a build of the
// ticked targets. Create it once.
func NewBuildTargetsPanel(e *env.Env, doctor *doctorCache, repoID func() string, build func()) *BuildTargetsPanel {
	p := &BuildTargetsPanel{env: e, repoID: repoID, doctor: doctor, boxes: map[string]*checkbox.Checkbox{}}
	for _, t := range project.Targets {
		id := t.ID
		p.boxes[id] = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
			OnChange(func(on bool) { p.set(id, on) })
	}
	p.recheck = components.NewIconAction("fetch", "Check again with giode doctor", func() { doctor.refresh(p.dir()) })
	p.build = components.Button("Build", "play", components.Primary, build)
	return p
}

func (p *BuildTargetsPanel) dir() string {
	if r := p.env.WS.Repo(p.repoID()); r != nil {
		return r.Path
	}
	return ""
}

// Targets is what Build builds for the repository: the saved choice, or
// this computer's own platform before anything was chosen.
func (p *BuildTargetsPanel) Targets() []string {
	if r := p.env.WS.Repo(p.repoID()); r != nil && len(r.BuildTargets) > 0 {
		return r.BuildTargets
	}
	return project.DefaultTargets()
}

func (p *BuildTargetsPanel) set(id string, on bool) {
	var next []string
	for _, t := range project.Targets {
		chosen := t.ID == id && on
		if t.ID != id {
			for _, cur := range p.Targets() {
				chosen = chosen || cur == t.ID
			}
		}
		if chosen {
			next = append(next, t.ID)
		}
	}
	if len(next) == 0 {
		p.env.Toasts.Show(giode.ToastInfo, "Keep at least one target")
		return
	}
	p.env.Fail("Saving build targets", p.env.WS.SetBuildTargets(p.repoID(), next))
}

func (p *BuildTargetsPanel) row(t project.Target, d project.Doctor, known bool) giode.Element {
	avail := project.Availability{OK: true}
	if known {
		avail = d.Available(t)
	}
	chosen := false
	for _, id := range p.Targets() {
		chosen = chosen || id == t.ID
	}
	p.boxes[t.ID].SetChecked(chosen)
	box := giode.Element(p.boxes[t.ID])
	title, sub := theme.Body, theme.CodeDim
	if !avail.OK && !chosen {
		box = giode.Box(giode.Styles{Width: 18, Height: 18, BorderRadius: 4, BorderWidth: 1, BorderColor: theme.Hairline, Background: theme.Sunken})
		title = theme.Tinted(title, theme.Faint)
	}
	right := giode.Element(giode.Text(t.ID, sub))
	if !avail.OK {
		right = giode.Text(avail.Reason, giode.Styles{FontSize: 11.5, Color: theme.Warning, MaxWidth: 240, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(2, 3)},
		box, giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(t.Title, title)), right)
}

func (p *BuildTargetsPanel) status(loaded bool, err error) giode.Element {
	switch {
	case !loaded:
		return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 12, Height: 12, Color: theme.Primary}), giode.Text("Asking giode doctor what this computer can build…", theme.Caption))
	case err != nil:
		msg := err.Error()
		if hint := project.InstallHint("giode"); strings.Contains(msg, "not found") {
			msg = hint
		}
		return giode.Text(msg, giode.Styles{FontSize: 12, Color: theme.Warning, MaxWidth: 420})
	}
	return giode.Text("Checked with giode doctor. Your choice is saved for this repository.", theme.Caption)
}

// Layout draws the panel.
func (p *BuildTargetsPanel) Layout(gtx giode.Context) giode.Dimensions {
	d, loaded, err := p.doctor.get(p.dir())
	rows := []giode.Element{
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Build targets", theme.Overline)), p.recheck),
		p.status(loaded, err),
		giode.Divider(giode.Styles{Color: theme.Hairline}),
	}
	for _, t := range project.Targets {
		rows = append(rows, p.row(t, d, loaded && err == nil))
	}
	p.build.Label("Build " + count(len(p.Targets()), "target", "targets"))
	rows = append(rows, giode.Divider(giode.Styles{Color: theme.Hairline}),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd}, p.build))
	return giode.Stack(giode.Styles{Width: 440, Gap: 6}, rows...).Layout(gtx)
}
