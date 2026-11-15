package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// TagsView lists tags and deletes any number of them at once, locally
// or on the server too, in a single push.
type TagsView struct {
	env       *env.Env
	repo      *core.RepoState
	tags      []gitx.Tag
	shown     []gitx.Tag
	sel       *components.Selection
	search    *components.SearchBox
	selectAll *checkbox.Checkbox
	list      *giode.LazyListElement
	bulk      *components.BulkBar
	create    *giode.ButtonElement
	pushAll   *giode.ButtonElement
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
		v.act("Pushing tags", "Tags pushed", gitx.PushTags)
	})
	v.list = giode.LazyList(
		func() int { return len(v.shown) },
		func(i int) any { return v.shown[i].Name },
		func(i int) giode.Element { return v.row(v.shown[i]) },
		giode.Styles{Gap: 2, ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	v.bulk = components.NewBulkBar(v.sel, "tags",
		components.Button("Delete locally", "trash", components.Secondary, func() { v.confirmDelete(false) }),
		components.Button("Delete everywhere", "trash", components.Destructive, func() { v.confirmDelete(true) }),
	)
	return v
}

// Open loads a repository's tags.
func (v *TagsView) Open(r *core.RepoState) {
	if v.repo == nil || v.repo.ID != r.ID {
		v.sel.Clear()
	}
	v.repo = r
	v.load()
}

func (v *TagsView) load() {
	path := v.repo.Path
	core.Query(v.env.WS, func(ctx context.Context) ([]gitx.Tag, error) {
		return gitx.Tags(ctx, path)
	}, func(tags []gitx.Tag, err error) {
		v.env.Fail("Loading tags", err)
		v.tags = tags
		names := make([]string, len(tags))
		for i, t := range tags {
			names[i] = t.Name
		}
		v.sel.Retain(names)
	})
}

func (v *TagsView) act(busy, success string, fn func(ctx context.Context, path string) error) {
	v.env.WS.Do(v.repo.ID, busy, fn, func(err error) {
		v.env.Done(success, err)
		v.load()
	})
}

func (v *TagsView) newTag() {
	v.prompt.Ask("New tag", "Tag name (on the current commit)", "v1.0.0", "", "Create tag", func(name string) {
		v.act("Tagging", "Tagged "+name, func(ctx context.Context, path string) error {
			return gitx.CreateTag(ctx, path, name, "", "")
		})
	})
}

func (v *TagsView) confirmDelete(remote bool) {
	names := v.sel.Keys()
	where, origin := "from this computer", ""
	if remote {
		where, origin = "from this computer and from origin, for everyone", "origin"
	}
	body := fmt.Sprintf("%d tags will be deleted %s.", len(names), where)
	v.confirm.Ask(fmt.Sprintf("Delete %d tags?", len(names)), body, "Delete", true, func() {
		v.sel.Clear()
		v.act("Deleting tags", fmt.Sprintf("Deleted %d tags", len(names)), func(ctx context.Context, path string) error {
			return gitx.DeleteTags(ctx, path, names, origin)
		})
	})
}

func (v *TagsView) row(t gitx.Tag) giode.Element {
	return components.SelectRow("tag-row", t.Name, v.sel, func() giode.Element {
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 180}, components.Pill("tag", t.Name, theme.Accent, theme.AccentSoft)),
			giode.Text(t.Hash, theme.CodeDim),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(t.Subject, theme.Ellipsis(theme.Caption))),
			giode.Text(components.Ago(t.When), theme.Caption),
			components.RowMenu("tag", t.Name,
				components.MenuItem{Icon: "copy", Label: "Copy name", Run: func() { v.env.Copy("Tag", t.Name) }},
				components.MenuItem{Icon: "trash", Label: "Delete everywhere", Run: func() {
					v.sel.Clear()
					v.sel.Set(t.Name, true)
					v.confirmDelete(true)
				}},
			),
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

// Layout draws the view.
func (v *TagsView) Layout(gtx giode.Context) giode.Dimensions {
	v.filter()
	list := giode.Element(giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(8)}, v.list))
	if len(v.shown) == 0 {
		list = components.EmptyState("tag", "No tags", "Tag a release to mark it in history.", nil)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 20}},
				v.selectAll,
				giode.Box(giode.Styles{Width: 320}, v.search),
				giode.Text(fmt.Sprintf("%d of %d", len(v.shown), len(v.tags)), theme.Caption),
				giode.Box(giode.Styles{FlexGrow: 1}), v.reload, v.pushAll, v.create),
			components.ListFrame(list),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 20, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, v.bulk),
		v.prompt, v.confirm,
	).Layout(gtx)
}
