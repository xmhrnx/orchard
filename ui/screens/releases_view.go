package screens

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// releasesData is a loaded release list and the service's features.
type releasesData struct {
	releases []forge.Release
	caps     forge.ReleaseCaps
}

// ReleasesView lists a repository's releases on its hosting service:
// read the notes, download assets, create, publish drafts, and delete
// many releases at once, with or without their tags.
type ReleasesView struct {
	env       *env.Env
	repo      *core.RepoState
	account   forge.Account
	slug      string
	linked    bool
	supported bool
	loading   bool
	loadErr   error
	data      releasesData
	shown     []forge.Release
	opened    string
	sel       *components.Selection
	search    *components.SearchBox
	selectAll *checkbox.Checkbox
	list      *giode.LazyListElement
	bulk      *components.BulkBar
	create    *giode.ButtonElement
	reload    *components.IconAction
	connect   *giode.ButtonElement
	retry     *giode.ButtonElement
	confirm   *components.ConfirmDialog
	dialog    *NewReleaseDialog
	panel     *ReleasePanel
}

// NewReleasesView creates the view. Create it once.
func NewReleasesView(e *env.Env) *ReleasesView {
	v := &ReleasesView{env: e, sel: components.NewSelection(), confirm: components.NewConfirmDialog()}
	v.search = components.NewSearchBox("Filter by tag or title")
	v.selectAll = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
		OnChange(func(on bool) { v.sel.SetAll(v.shownIDs(), on) })
	v.reload = components.NewIconAction("fetch", "Reload", v.load)
	v.retry = components.Button("Try again", "fetch", components.Secondary, v.load)
	v.connect = components.Button("Connect an account", "user", components.Primary, func() { e.Nav.Open(env.PageAccounts) })
	v.dialog = NewNewReleaseDialog(e, v.load)
	v.create = components.Button("New release", "plus", components.Primary, func() {
		v.dialog.Open(v.account, v.slug, v.repo.Status.Branch, v.data.caps)
	})
	v.panel = NewReleasePanel(e, ReleaseActions{
		Publish: v.publish,
		Delete:  func(r forge.Release) { v.confirmDelete([]forge.Release{r}, false) },
		Close:   func() { v.opened = "" },
	})
	v.list = giode.LazyList(
		func() int { return len(v.shown) },
		func(i int) any { return v.shown[i].ID },
		func(i int) giode.Element { return v.row(v.shown[i]) },
		giode.Styles{Gap: 2, ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	v.bulk = components.NewBulkBar(v.sel, "release", "releases",
		components.Button("Delete", "trash", components.Secondary, func() { v.confirmDelete(v.selected(), false) }),
		components.Button("Delete with tags", "trash", components.Destructive, func() { v.confirmDelete(v.selected(), true) }),
	)
	return v
}

// Open loads the releases of a repository.
func (v *ReleasesView) Open(r *core.RepoState) {
	if v.repo == nil || v.repo.ID != r.ID {
		v.sel.Clear()
		v.data = releasesData{}
		v.opened = ""
	}
	v.repo = r
	v.account, v.slug, v.linked = v.env.WS.HostedRepo(r)
	v.supported = true
	v.load()
}

func (v *ReleasesView) load() {
	if !v.linked {
		return
	}
	acc, slug, repoID := v.account, v.slug, v.repo.ID
	v.loading = true
	core.Query(v.env.WS, func(ctx context.Context) (releasesData, error) {
		client, err := v.env.WS.Client(acc)
		if err != nil {
			return releasesData{}, err
		}
		rs, ok := forge.ReleasesOf(client)
		if !ok {
			return releasesData{}, forge.ErrNoReleases
		}
		list, err := rs.Releases(ctx, slug)
		return releasesData{releases: list, caps: rs.ReleaseCaps()}, err
	}, func(d releasesData, err error) {
		if v.repo == nil || v.repo.ID != repoID {
			return
		}
		v.loading, v.loadErr = false, nil
		if errors.Is(err, forge.ErrNoReleases) {
			v.supported = false
			return
		}
		if err != nil {
			v.loadErr = err
			return
		}
		v.data = d
		v.sel.Retain(v.allIDs())
		v.refreshPanel()
	})
}

// refreshPanel shows the reloaded version of the open release, or
// closes the panel when it is gone.
func (v *ReleasesView) refreshPanel() {
	for _, r := range v.data.releases {
		if r.ID == v.opened {
			v.panel.Show(r)
			return
		}
	}
	v.opened = ""
}

func (v *ReleasesView) selected() []forge.Release {
	var out []forge.Release
	for _, r := range v.data.releases {
		if v.sel.Has(r.ID) {
			out = append(out, r)
		}
	}
	return out
}

func (v *ReleasesView) publish(r forge.Release) {
	slug := v.slug
	forgeBulk(v.env, v.account, []forge.Release{r}, releaseTitle,
		func(ctx context.Context, c forge.Client, r forge.Release) error {
			rs, ok := forge.ReleasesOf(c)
			if !ok {
				return forge.ErrNoReleases
			}
			return rs.PublishRelease(ctx, slug, r.ID)
		}, func(ok int, err error) {
			if ok > 0 {
				v.env.Toasts.Show(giode.ToastSuccess, "Published "+releaseTitle(r))
			}
			v.env.Fail("Publishing", err)
			v.load()
		})
}

func (v *ReleasesView) confirmDelete(releases []forge.Release, withTags bool) {
	if len(releases) == 0 {
		return
	}
	n := count(len(releases), "release", "releases")
	body := fmt.Sprintf("%s with notes and assets will be deleted from %s. Tags stay.", n, v.account.Host)
	verb := "Delete"
	if withTags {
		body = fmt.Sprintf("%s will be deleted from %s, and their tags removed from origin and this computer.", n, v.account.Host)
		verb = "Delete all"
	}
	v.confirm.Ask("Delete "+n+"?", body, verb, true, func() {
		v.sel.Clear()
		v.delete(releases, withTags)
	})
}

func (v *ReleasesView) delete(releases []forge.Release, withTags bool) {
	slug, repoID := v.slug, v.repo.ID
	forgeBulk(v.env, v.account, releases, releaseTitle,
		func(ctx context.Context, c forge.Client, r forge.Release) error {
			rs, ok := forge.ReleasesOf(c)
			if !ok {
				return forge.ErrNoReleases
			}
			return rs.DeleteRelease(ctx, slug, r.ID)
		}, func(ok int, err error) {
			if ok > 0 {
				v.env.Toasts.Show(giode.ToastSuccess, "Deleted "+count(ok, "release", "releases"))
			}
			v.env.Fail("Some releases weren't deleted", err)
			if withTags && ok > 0 {
				v.deleteTags(repoID, releases)
			}
			v.load()
		})
}

// deleteTags removes the tags of deleted releases with git.
func (v *ReleasesView) deleteTags(repoID string, releases []forge.Release) {
	tags := make([]string, len(releases))
	for i, r := range releases {
		tags[i] = r.Tag
	}
	v.env.WS.Do(repoID, "Deleting tags", func(ctx context.Context, path string) error {
		return gitx.DeleteTagsIfPresent(ctx, path, tags)
	}, func(err error) { v.env.Done("Removed "+count(len(tags), "tag", "tags"), err) })
}

func (v *ReleasesView) row(r forge.Release) giode.Element {
	show := giode.Keyed("release-open", r.ID, func() *components.IconAction {
		id := r.ID
		return components.NewIconAction("chevron-right", "Show release", func() {
			for _, rel := range v.data.releases {
				if rel.ID == id {
					v.opened = id
					v.panel.Show(rel)
				}
			}
		})
	})
	return components.SelectRow("release-row", r.ID, v.sel, func() giode.Element {
		badges := []giode.Element{giode.Text(releaseTitle(r), theme.Ellipsis(theme.Strong))}
		switch {
		case r.Latest:
			badges = append(badges, components.Pill("star", "latest", theme.OnPrimary, theme.Primary))
		case r.Draft:
			badges = append(badges, components.Pill("", "draft", theme.Muted, theme.Hover))
		case r.Prerelease:
			badges = append(badges, components.Pill("", "pre-release", theme.Warning, theme.AccentSoft))
		}
		meta := []string{}
		if r.Author != "" {
			meta = append(meta, r.Author)
		}
		if n := len(r.Assets); n > 0 {
			meta = append(meta, fmt.Sprintf("%d assets", n))
		}
		when := r.Published
		if when.IsZero() {
			when = r.Created
		}
		meta = append(meta, components.Ago(when))
		bg := giode.Color("transparent")
		if v.opened == r.ID {
			bg = theme.Hover
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Background: bg, BorderRadius: 8},
			giode.Box(giode.Styles{Padding: giode.UniformInset(7), BorderRadius: 9, Background: theme.PrimarySoft}, components.Glyph("release", 16, theme.Primary)),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, badges...),
				giode.Text(strings.Join(meta, " · "), theme.Ellipsis(theme.Caption))),
			components.Pill("tag", r.Tag, theme.Accent, theme.AccentSoft),
			show,
		)
	})
}

func (v *ReleasesView) allIDs() []string {
	ids := make([]string, len(v.data.releases))
	for i, r := range v.data.releases {
		ids[i] = r.ID
	}
	return ids
}

func (v *ReleasesView) shownIDs() []string {
	ids := make([]string, len(v.shown))
	for i, r := range v.shown {
		ids[i] = r.ID
	}
	return ids
}

func (v *ReleasesView) filter() {
	q := strings.ToLower(strings.TrimSpace(v.search.Query()))
	v.shown = v.shown[:0]
	for _, r := range v.data.releases {
		if q == "" || strings.Contains(strings.ToLower(r.Tag+" "+r.Name), q) {
			v.shown = append(v.shown, r)
		}
	}
	v.selectAll.SetChecked(v.sel.AllOf(v.shownIDs()))
}

func (v *ReleasesView) body() giode.Element {
	switch {
	case !v.linked:
		host := "this repository's host"
		if loc, ok := forge.ParseRemote(v.repo.Origin); ok {
			host = loc.Host
		}
		return components.EmptyState("release", "Connect "+host,
			"Releases live on the hosting service. Connect an account for "+host+" to publish and tidy them here.", v.connect)
	case !v.supported:
		return components.EmptyState("release", "No releases here", v.account.Kind.Title()+" doesn't have releases. Tags still work from the Tags tab.", nil)
	case v.loadErr != nil:
		return components.EmptyState("alert", "Couldn't load releases", v.loadErr.Error(), v.retry)
	case v.loading && len(v.data.releases) == 0:
		return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(40), Justify: giode.JustifyCenter},
			giode.Spinner(giode.Styles{Width: 28, Height: 28, Color: theme.Primary}))
	case len(v.data.releases) == 0:
		return components.EmptyState("release", "No releases yet", "Publish the first one to give people something to download.", v.create)
	case len(v.shown) == 0:
		return components.EmptyState("filter", "Nothing matches", "Try another filter.", nil)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(8)}, v.list)
}

// Layout draws the view.
func (v *ReleasesView) Layout(gtx giode.Context) giode.Dimensions {
	v.filter()
	var head []giode.Element
	if v.linked && v.supported {
		head = []giode.Element{v.selectAll, giode.Box(giode.Styles{Width: 320}, v.search),
			giode.Text(fmt.Sprintf("%d releases", len(v.data.releases)), theme.Caption),
			giode.Box(giode.Styles{FlexGrow: 1}), v.reload, v.create}
	} else {
		head = []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(v.slug, theme.CodeDim))}
	}
	content := []giode.Element{components.FilledFrame(giode.Styles{FlexGrow: 1, Height: -1, Padding: giode.UniformInset(0)}, v.body())}
	if v.opened != "" {
		content = append(content, v.panel)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 20}}, head...),
			giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Gap: 14, Wrap: giode.WrapNoWrap}, content...),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 20, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, v.bulk),
		v.confirm, v.dialog,
	).Layout(gtx)
}
