package screens

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/xmhrnx/giode"

	"orchard/internal/core"
	"orchard/internal/project"
	"orchard/internal/tasks"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// Task kinds, so starting one again replaces the one running.
const (
	taskRun   = "run"
	taskDev   = "dev"
	taskBuild = "build"
	taskTool  = "tool"
)

// detected is a repository's project, once looked for.
type detected struct {
	project project.Project
	ok      bool
}

// ProjectBar puts the project's own commands in the repository header:
// Run and a Build with a target dropdown for Giode apps, the start, dev
// and build scripts for npm projects, run and build for Go modules.
type ProjectBar struct {
	env      *env.Env
	repoID   func() string
	found    map[string]detected
	doctor   *doctorCache
	targets  *BuildTargetsPanel
	popover  *components.Popover
	run      *giode.ButtonElement
	dev      *giode.ButtonElement
	build    *giode.ButtonElement
	chooser  *components.IconAction
	giode    *giode.MenuElement
	giodeBtn giode.Element
	node     *giode.MenuElement
	nodeBtn  giode.Element
}

// NewProjectBar creates the bar for whichever repository repoID names.
// Create it once.
func NewProjectBar(e *env.Env, repoID func() string) *ProjectBar {
	b := &ProjectBar{env: e, repoID: repoID, found: map[string]detected{}, doctor: newDoctorCache(e)}
	b.targets = NewBuildTargetsPanel(e, b.doctor, repoID, func() {
		b.popover.Close()
		b.startBuild()
	})
	b.popover = components.NewPopover(func() giode.Element { return b.targets })
	b.run = components.Button("Run", "play", components.Primary, func() { b.toggle(taskRun) })
	b.dev = components.Button("Dev", "terminal", components.Secondary, func() { b.toggle(taskDev) })
	b.build = components.Button("Build", "release", components.Secondary, b.startBuild)
	b.chooser = components.NewIconAction("chevron-down", "Choose build targets", b.popover.Toggle)
	b.giode = giode.Menu().
		IconItem(components.GlyphSource("sparkle"), "Doctor: what can this computer build?", func() { b.tool("Doctor", "doctor") }).
		Separator().
		IconItem(components.GlyphSource("arrow-up"), "Upgrade Giode in this project", func() { b.tool("Upgrade Giode", "upgrade") }).
		IconItem(components.GlyphSource("download"), "Update the giode CLI", func() { b.tool("Update giode CLI", "self-update") }).
		Separator().
		IconItem(components.GlyphSource("tag"), "Bump patch version", func() { b.tool("Bump patch", "bump", "patch") }).
		IconItem(components.GlyphSource("tag"), "Bump minor version", func() { b.tool("Bump minor", "bump", "minor") }).
		IconItem(components.GlyphSource("tag"), "Bump major version", func() { b.tool("Bump major", "bump", "major") }).
		Separator().
		IconItem(components.GlyphSource("folder"), "Open build output", b.openOutput).
		IconItem(components.GlyphSource("eraser"), "Clean build output", func() { b.tool("Clean", "clean") })
	b.giodeBtn = b.giode.Trigger(components.NewIconAction("dots", "Giode tools", func() {}))
	b.node = giode.Menu().
		IconItem(components.GlyphSource("download"), "Install dependencies", b.install)
	b.nodeBtn = b.node.Trigger(components.NewIconAction("dots", "Package tools", func() {}))
	return b
}

// Detect looks for the project of a repository, once per opening.
func (b *ProjectBar) Detect(r *core.RepoState) {
	id, path := r.ID, r.Path
	core.Query(b.env.WS, func(context.Context) (detected, error) {
		p, ok, err := project.Detect(path)
		return detected{p, ok}, err
	}, func(d detected, err error) {
		b.env.Fail("Reading the project", err)
		b.found[id] = d
	})
}

func (b *ProjectBar) current() (project.Project, bool) {
	d, ok := b.found[b.repoID()]
	return d.project, ok && d.ok
}

func (b *ProjectBar) start(kind, title string, steps []project.Command) {
	r := b.env.WS.Repo(b.repoID())
	if r == nil || len(steps) == 0 {
		return
	}
	b.env.Tasks.Start(r.ID, r.Path, kind, title, steps)
}

// toggle starts a run or dev task, or stops it when it's running.
func (b *ProjectBar) toggle(kind string) {
	if t := b.env.Tasks.Running(b.repoID(), kind); t != nil {
		t.Stop()
		return
	}
	p, ok := b.current()
	if !ok {
		return
	}
	switch {
	case p.Kind == project.Giode:
		b.start(kind, "Run", project.GiodeRun())
	case p.Kind == project.Go:
		b.start(kind, "Run", project.GoRun())
	case kind == taskDev:
		b.start(kind, "Dev", p.NodeScript("dev"))
	default:
		b.start(kind, "Run", p.NodeScript("start"))
	}
}

func (b *ProjectBar) startBuild() {
	p, ok := b.current()
	if !ok {
		return
	}
	switch p.Kind {
	case project.Giode:
		targets := b.targets.Targets()
		b.start(taskBuild, "Build "+strings.Join(targets, ", "), project.GiodeBuild(targets))
	case project.Go:
		b.start(taskBuild, "Build", project.GoBuild())
	default:
		b.start(taskBuild, "Build", p.NodeScript("build"))
	}
}

func (b *ProjectBar) tool(title string, args ...string) {
	b.start(taskTool, title, project.GiodeTool(args...))
}

func (b *ProjectBar) install() {
	if p, ok := b.current(); ok {
		b.start(taskTool, "Install", p.NodeInstall())
	}
}

func (b *ProjectBar) openOutput() {
	if r := b.env.WS.Repo(b.repoID()); r != nil {
		b.env.Reveal(filepath.Join(r.Path, "dist"))
	}
}

// running reports whether a task of kind runs for the repository.
func (b *ProjectBar) running(kind string) bool {
	t := b.env.Tasks.Running(b.repoID(), kind)
	return t != nil && t.State() == tasks.Running
}

// runButton shows Run, or Stop while the task runs.
func (b *ProjectBar) runButton(btn *giode.ButtonElement, kind, label string) giode.Element {
	if b.running(kind) {
		stop := "Stop"
		if label != "Run" {
			stop = "Stop " + strings.ToLower(label)
		}
		btn.Label(stop).Icon(components.Glyph("stop", 15, ""))
	} else {
		btn.Label(label).Icon(components.Glyph("play", 15, ""))
	}
	return btn
}

// Layout draws the bar; it is empty for repositories that aren't a
// recognised project.
func (b *ProjectBar) Layout(gtx giode.Context) giode.Dimensions {
	p, ok := b.current()
	if !ok {
		return giode.Box(giode.Styles{}).Layout(gtx)
	}
	b.build.Loading(b.running(taskBuild))
	var kids []giode.Element
	switch p.Kind {
	case project.Giode:
		b.build.Label("Build")
		kids = []giode.Element{
			b.runButton(b.run, taskRun, "Run"),
			giode.Row(giode.Styles{Gap: 0, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, b.build, b.popover.Anchor(b.chooser)),
			b.giodeBtn, b.giode,
		}
	case project.Go:
		kids = []giode.Element{b.runButton(b.run, taskRun, "Run"), b.build}
	case project.Node:
		if p.HasScript("start") {
			kids = append(kids, b.runButton(b.run, taskRun, "Run"))
		}
		if p.HasScript("dev") {
			kids = append(kids, b.runButton(b.dev, taskDev, "Dev"))
		}
		if p.HasScript("build") {
			kids = append(kids, b.build)
		}
		kids = append(kids, b.nodeBtn, b.node)
	}
	badge := components.Pill("", p.Kind.Title(), theme.Muted, theme.Hover)
	if p.Kind == project.Node {
		badge = components.Pill("", p.Manager, theme.Muted, theme.Hover)
	}
	return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		append([]giode.Element{badge}, kids...)...).Layout(gtx)
}
