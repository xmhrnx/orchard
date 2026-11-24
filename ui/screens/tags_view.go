package screens

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// tagEntry is a tag and where it exists: here, on origin, or both.
// Differs marks a name used on both sides for different commits.
type tagEntry struct {
	gitx.Tag
	Local   bool
	Remote  bool
	Differs bool
}

// TagsView lists the tags here and on origin side by side, so a tag
// missing on either side shows, and deletes any number of them at once,
// locally or on the server too, in a single push.
type TagsView struct {
	env       *env.Env
	repo      *core.RepoState
	local     []gitx.Tag
	remote    []gitx.RemoteTag
	remoteErr error
	checking  bool
	synced    time.Time
	tags      []tagEntry
	shown     []tagEntry
	sel       *components.Selection
	search    *components.SearchBox
	selectAll *checkbox.Checkbox
	list      *giode.LazyListElement
	bulk      *components.BulkBar
	create    *giode.ButtonElement
	pushAll   *giode.ButtonElement
	fetchAll  *giode.ButtonElement
	reload    *components.IconAction
	prompt    *components.PromptDialog
	confirm   *components.ConfirmDialog
}

// NewTagsView creates the view. Create it once.
func NewTagsView(e *env.Env) *TagsView {
	v := &TagsView{env: e, sel: components.NewSelection()}
	v.search = components.NewSearchBox("Filter tags, e.g. v1. or -rc")
	v.selectAll = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
		OnChange(func(on bool) { v.sel.SetAll(v.shownNames(), on) })
	v.prompt = components.NewPromptDialog()
	v.confirm = components.NewConfirmDialog()
	v.reload = components.NewIconAction("fetch", "Reload", v.load)
	v.create = components.Button("New tag", "plus", components.Primary, v.newTag)
	v.pushAll = components.Button("Push tags", "push", components.Secondary, func() {
		v.act("Pushing tags", "Tags pushed to origin", gitx.PushTags)
	})
	v.fetchAll = components.Button("Fetch tags", "pull", components.Secondary, func() {
		v.act("Fetching tags", "Tags fetched from origin", func(ctx context.Context, path string) error {
			return gitx.FetchTags(ctx, path, "origin")
		})
	})
	v.list = giode.LazyList(
		func() int { return len(v.shown) },
		func(i int) any { return v.shown[i].Name },
		func(i int) giode.Element { return v.row(v.shown[i]) },
		giode.Styles{Gap: 2, ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	v.bulk = components.NewBulkBar(v.sel, "tag", "tags",
		components.Button("Delete locally", "trash", components.Secondary, func() { v.confirmDelete(false) }),
		components.Button("Delete everywhere", "trash", components.Destructive, func() { v.confirmDelete(true) }),
	)
	return v
}

// Open loads a repository's tags.
func (v *TagsView) Open(r *core.RepoState) {
	if v.repo == nil || v.repo.ID != r.ID {
		v.sel.Clear()
		v.local, v.remote, v.tags = nil, nil, nil
	}
	v.repo = r
	v.load()
}

// hasOrigin reports whether there is a server to compare with.
func (v *TagsView) hasOrigin() bool { return v.repo != nil && v.repo.Origin != "" }

// load reads the local tags at once and origin's in the background,
// since asking the server takes a moment.
func (v *TagsView) load() {
	path, repoID := v.repo.Path, v.repo.ID
	v.synced = v.repo.Synced
	core.Query(v.env.WS, func(ctx context.Context) ([]gitx.Tag, error) {
		return gitx.Tags(ctx, path)
	}, func(tags []gitx.Tag, err error) {
		if v.repo.ID != repoID {
			return
		}
		v.env.Fail("Loading tags", err)
		v.local = tags
		v.merge()
	})
	if !v.hasOrigin() {
		v.remote, v.remoteErr = nil, nil
		return
	}
	v.checking = true
	core.Query(v.env.WS, func(ctx context.Context) ([]gitx.RemoteTag, error) {
		return gitx.RemoteTags(ctx, path, "origin")
	}, func(tags []gitx.RemoteTag, err error) {
		if v.repo.ID != repoID {
			return
		}
		v.checking, v.remoteErr = false, err
		if err == nil {
			v.remote = tags
		}
		v.merge()
	})
}

// compared reports whether origin's tags are known.
func (v *TagsView) compared() bool { return v.hasOrigin() && v.remoteErr == nil && v.remote != nil }

// merge combines local and origin tags: tags only on origin come first,
// then local tags newest first.
func (v *TagsView) merge() {
	onOrigin := map[string]gitx.RemoteTag{}
	for _, r := range v.remote {
		onOrigin[r.Name] = r
	}
	here := map[string]bool{}
	var entries []tagEntry
	for _, t := range v.local {
		here[t.Name] = true
		r, remote := onOrigin[t.Name]
		e := tagEntry{Tag: t, Local: true, Remote: remote || !v.compared()}
		e.Differs = remote && r.Commit != t.Hash
		entries = append(entries, e)
	}
	var remoteOnly []tagEntry
	for _, r := range v.remote {
		if !here[r.Name] {
			remoteOnly = append(remoteOnly, tagEntry{Tag: gitx.Tag{Name: r.Name, Hash: r.Commit}, Remote: true})
		}
	}
	sort.Slice(remoteOnly, func(i, j int) bool { return remoteOnly[i].Name > remoteOnly[j].Name })
	v.tags = append(remoteOnly, entries...)
	names := make([]string, len(v.tags))
	for i, t := range v.tags {
		names[i] = t.Name
	}
	v.sel.Retain(names)
}

func (v *TagsView) act(busy, success string, fn func(ctx context.Context, path string) error) {
	v.env.WS.Do(v.repo.ID, busy, fn, func(err error) {
		v.env.Done(success, err)
		v.load()
	})
}

func (v *TagsView) newTag() {
	label := "Tag name (on the current commit, pushed to origin too)"
	if !v.hasOrigin() {
		label = "Tag name (on the current commit)"
	}
	v.prompt.Ask("New tag", label, "v1.0.0", "", "Create tag", func(name string) {
		pushed := false
		v.env.WS.Do(v.repo.ID, "Tagging", func(ctx context.Context, path string) error {
			var err error
			pushed, err = createAndPushTag(ctx, path, name, "")
			return err
		}, func(err error) {
			v.env.Done(tagCreated(name, pushed), err)
			v.load()
		})
	})
}

func (v *TagsView) confirmDelete(everywhere bool) {
	var names, local []string
	for _, t := range v.tags {
		if v.sel.Has(t.Name) {
			names = append(names, t.Name)
			if t.Local {
				local = append(local, t.Name)
			}
		}
	}
	if !everywhere {
		if len(local) == 0 {
			v.env.Toasts.Show(giode.ToastInfo, "None of these tags exist on this computer")
			return
		}
		n := count(len(local), "tag", "tags")
		v.confirm.Ask("Delete "+n+"?", n+" will be deleted from this computer. Origin keeps its copies.", "Delete", true, func() {
			v.sel.Clear()
			v.act("Deleting tags", "Deleted "+n, func(ctx context.Context, path string) error {
				return gitx.DeleteTags(ctx, path, local, "")
			})
		})
		return
	}
	if v.hasOrigin() && !v.compared() {
		v.env.Toasts.Show(giode.ToastWarning, "Orchard couldn't check origin's tags yet, so it won't delete there. Reload and try again.")
		return
	}
	plan := gitx.PlanTagDeletions(v.local, v.remote, names, gitx.KeepOrigin)
	v.confirm.Ask("Delete "+count(len(plan), "tag", "tags")+"?", deletionSummary(plan, v.differing(names)), "Delete", true, func() {
		v.sel.Clear()
		v.act("Deleting tags", "Deleted "+count(len(plan), "tag", "tags"), func(ctx context.Context, path string) error {
			return gitx.DeleteTagsSafely(ctx, path, "origin", plan)
		})
	})
}

// differing counts the named tags that differ between here and origin.
func (v *TagsView) differing(names []string) int {
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}
	n := 0
	for _, t := range v.tags {
		if want[t.Name] && t.Differs {
			n++
		}
	}
	return n
}

// deletionSummary says exactly what a deletion will remove where.
func deletionSummary(plan []gitx.TagDeletion, differing int) string {
	var both, here, there int
	for _, d := range plan {
		switch {
		case d.Local && d.RemoteObject != "":
			both++
		case d.Local:
			here++
		default:
			there++
		}
	}
	var parts []string
	if both > 0 {
		parts = append(parts, count(both, "tag", "tags")+" from this computer and from origin")
	}
	if here > 0 {
		parts = append(parts, count(here, "tag", "tags")+" from this computer only")
	}
	if there > 0 {
		parts = append(parts, count(there, "tag", "tags")+" from origin only")
	}
	body := "Deletes " + strings.Join(parts, ", ") + "."
	if differing > 0 {
		body += " " + count(differing, "tag points", "tags point") + " at a different commit on origin: origin's copy is kept."
	}
	return body
}

func (v *TagsView) where(t tagEntry) giode.Element {
	switch {
	case t.Differs:
		return components.Pill("alert", "differs from origin", theme.Danger, theme.DangerSoft)
	case !t.Local:
		return components.Pill("globe", "only on origin", theme.Info, theme.InfoSoft)
	case !t.Remote:
		return components.Pill("push", "not on origin", theme.Warning, theme.AccentSoft)
	case !t.Annotated:
		return components.Pill("", "lightweight", theme.Muted, theme.Hover)
	}
	return giode.Box(giode.Styles{})
}

func (v *TagsView) menu(t tagEntry) giode.Element {
	return components.RowMenu("tag", t.Name,
		components.MenuItem{Icon: "copy", Label: "Copy name", Run: func() { v.env.Copy("Tag", t.Name) }},
		components.MenuItem{Icon: "push", Label: "Push to origin", Run: func() {
			v.act("Pushing "+t.Name, "Pushed "+t.Name, func(ctx context.Context, path string) error {
				return gitx.PushTag(ctx, path, "origin", t.Name)
			})
		}},
		components.MenuItem{Icon: "pull", Label: "Fetch from origin", Run: func() {
			v.act("Fetching tags", "Fetched tags from origin", func(ctx context.Context, path string) error {
				return gitx.FetchTags(ctx, path, "origin")
			})
		}},
		components.MenuSeparator,
		components.MenuItem{Icon: "trash", Label: "Delete everywhere", Run: func() {
			v.sel.Clear()
			v.sel.Set(t.Name, true)
			v.confirmDelete(true)
		}},
	)
}

func (v *TagsView) row(t tagEntry) giode.Element {
	return components.SelectRow("tag-row", t.Name, v.sel, func() giode.Element {
		subject, age := t.Subject, components.Ago(t.When)
		if !t.Local {
			subject, age = "Fetch tags to bring it to this computer", ""
		} else if !t.Annotated && age != "" {
			age = "commit " + age
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 180}, components.Pill("tag", t.Name, theme.Accent, theme.AccentSoft)),
			giode.Text(t.Hash[:min(7, len(t.Hash))], theme.CodeDim),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(subject, theme.Ellipsis(theme.Caption))),
			v.where(t),
			giode.Box(giode.Styles{Width: 96}, giode.Text(age, giode.Styles{FontSize: 12, Color: theme.Muted, TextAlign: giode.TextAlignEnd, Width: -1})),
			v.menu(t),
		)
	})
}

func (v *TagsView) shownNames() []string {
	names := make([]string, len(v.shown))
	for i, t := range v.shown {
		names[i] = t.Name
	}
	return names
}

func (v *TagsView) filter() {
	q := strings.ToLower(v.search.Query())
	v.shown = v.shown[:0]
	for _, t := range v.tags {
		if q == "" || strings.Contains(strings.ToLower(t.Name), q) {
			v.shown = append(v.shown, t)
		}
	}
	v.selectAll.SetChecked(v.sel.AllOf(v.shownNames()))
}

// status says how the list compares with origin.
func (v *TagsView) status() giode.Element {
	text := fmt.Sprintf("%d of %d", len(v.shown), len(v.tags))
	switch {
	case !v.hasOrigin():
		text += " · no origin to compare with"
	case v.checking:
		return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Text(text, theme.Caption), giode.Spinner(giode.Styles{Width: 12, Height: 12, Color: theme.Muted}),
			giode.Text("checking origin…", theme.Caption))
	case v.remoteErr != nil:
		return giode.Text(text+" · couldn't check origin: "+v.remoteErr.Error(), theme.Ellipsis(theme.Tinted(theme.Caption, theme.Warning)))
	}
	return giode.Text(text, theme.Caption)
}

// Layout draws the view.
func (v *TagsView) Layout(gtx giode.Context) giode.Dimensions {
	if v.repo != nil && !v.repo.Synced.Equal(v.synced) {
		v.load() // a fetch or pull just ran: show what it brought in
	}
	v.filter()
	list := giode.Element(giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(8)}, v.list))
	if len(v.shown) == 0 {
		list = components.EmptyState("tag", "No tags", "Not here and not on origin. Tag a release to mark it in history.", nil)
	}
	actions := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, v.status()), v.reload}
	if v.hasOrigin() {
		actions = append(actions, v.fetchAll, v.pushAll)
	}
	actions = append(actions, v.create)
	head := append([]giode.Element{v.selectAll, giode.Box(giode.Styles{Width: 300}, v.search)}, actions...)
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 20}}, head...),
			components.ListFrame(list),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 20, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, v.bulk),
		v.prompt, v.confirm,
	).Layout(gtx)
}
