package screens

import (
	"context"
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// PullsView lists a repository's open pull (or merge) requests on its
// hosting service and closes or merges many at once.
type PullsView struct {
	env       *env.Env
	repo      *core.RepoState
	account   forge.Account
	slug      string
	linked    bool
	loading   bool
	pulls     []forge.PullRequest
	sel       *components.Selection
	selectAll *checkbox.Checkbox
	list      *giode.LazyListElement
	bulk      *components.BulkBar
	reload    *components.IconAction
	connect   *giode.ButtonElement
	confirm   *components.ConfirmDialog
}

// NewPullsView creates the view. Create it once.
func NewPullsView(e *env.Env) *PullsView {
	v := &PullsView{env: e, sel: components.NewSelection(), confirm: components.NewConfirmDialog()}
	v.selectAll = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
		OnChange(func(on bool) { v.sel.SetAll(v.keys(), on) })
	v.reload = components.NewIconAction("fetch", "Reload", v.load)
	v.connect = components.Button("Connect an account", "user", components.Primary, func() { e.Nav.Open(env.PageAccounts) })
	v.list = giode.LazyList(
		func() int { return len(v.pulls) },
		func(i int) any { return v.pulls[i].Number },
		func(i int) giode.Element { return v.row(v.pulls[i]) },
		giode.Styles{Gap: 2, ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	v.bulk = components.NewBulkBar(v.sel, "requests",
		components.Button("Merge", "merge", components.Secondary, func() { v.confirmBulk(true) }),
		components.Button("Close", "x", components.Destructive, func() { v.confirmBulk(false) }),
	)
	return v
}

func prKey(n int) string { return fmt.Sprint(n) }

func (v *PullsView) keys() []string {
	keys := make([]string, len(v.pulls))
	for i, p := range v.pulls {
		keys[i] = prKey(p.Number)
	}
	return keys
}

// Open loads the open requests of a repository.
func (v *PullsView) Open(r *core.RepoState) {
	if v.repo == nil || v.repo.ID != r.ID {
		v.sel.Clear()
		v.pulls = nil
	}
	v.repo = r
	v.account, v.slug, v.linked = v.env.WS.HostedRepo(r)
	v.load()
}

func (v *PullsView) load() {
	if !v.linked {
		return
	}
	acc, slug := v.account, v.slug
	v.loading = true
	core.Query(v.env.WS, func(ctx context.Context) ([]forge.PullRequest, error) {
		client, err := v.env.WS.Client(acc)
		if err != nil {
			return nil, err
		}
		return client.PullRequests(ctx, slug)
	}, func(list []forge.PullRequest, err error) {
		v.loading = false
		v.env.Fail("Loading pull requests", err)
		v.pulls = list
		v.sel.Retain(v.keys())
	})
}

func (v *PullsView) confirmBulk(merge bool) {
	var numbers []int
	for _, p := range v.pulls {
		if v.sel.Has(prKey(p.Number)) {
			numbers = append(numbers, p.Number)
		}
	}
	verb, body := "Close", "They will be closed without merging. They can be reopened on the website."
	if merge {
		verb, body = "Merge", "Each will be merged into its target branch with the service's default merge method."
	}
	v.confirm.Ask(fmt.Sprintf("%s %d requests?", verb, len(numbers)), body, verb, !merge, func() {
		v.sel.Clear()
		v.apply(numbers, merge)
	})
}

// apply closes or merges requests a few at a time.
func (v *PullsView) apply(numbers []int, merge bool) {
	verb, slug := "Closed", v.slug
	if merge {
		verb = "Merged"
	}
	forgeBulk(v.env, v.account, numbers, func(n int) string { return fmt.Sprintf("#%d", n) },
		func(ctx context.Context, c forge.Client, n int) error {
			if merge {
				return c.MergePullRequest(ctx, slug, n)
			}
			return c.ClosePullRequest(ctx, slug, n)
		}, func(ok int, err error) {
			if ok > 0 {
				v.env.Toasts.Show(giode.ToastSuccess, fmt.Sprintf("%s %d requests", verb, ok))
			}
			v.env.Fail("Some requests failed", err)
			v.load()
		})
}

func (v *PullsView) row(p forge.PullRequest) giode.Element {
	return components.SelectRow("pr-row", prKey(p.Number), v.sel, func() giode.Element {
		title := []giode.Element{
			giode.Text(fmt.Sprintf("#%d", p.Number), giode.Styles{FontSize: 12.5, FontFamily: theme.Mono, Color: theme.Muted}),
			giode.Text(p.Title, theme.Ellipsis(theme.Strong)),
		}
		if p.Draft {
			title = append(title, components.Pill("", "draft", theme.Muted, theme.Hover))
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			components.Glyph("pull-request", 16, theme.Primary),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, title...),
				giode.Text(fmt.Sprintf("%s → %s · by %s · %s", p.Source, p.Target, p.Author, components.Ago(p.Updated)), theme.Ellipsis(theme.Caption)),
			),
			components.RowMenu("pr", prKey(p.Number),
				components.MenuItem{Icon: "external", Label: "Open in browser", Run: func() { v.env.OpenURL(p.WebURL) }},
				components.MenuItem{Icon: "copy", Label: "Copy link", Run: func() { v.env.Copy("Link", p.WebURL) }},
			),
		)
	})
}

func (v *PullsView) body() giode.Element {
	switch {
	case !v.linked:
		host := "this repository's host"
		if loc, ok := forge.ParseRemote(v.repo.Origin); ok {
			host = loc.Host
		}
		return components.EmptyState("pull-request", "Connect "+host,
			"Pull and merge requests come from the hosting service. Connect an account for "+host+" to list, merge and close them here.", v.connect)
	case v.loading && len(v.pulls) == 0:
		return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(40), Justify: giode.JustifyCenter},
			giode.Spinner(giode.Styles{Width: 28, Height: 28, Color: theme.Primary}))
	case len(v.pulls) == 0:
		return components.EmptyState("check", "No open requests", "Everything has been reviewed. Nice and tidy.", nil)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(8)}, v.list)
}

// Layout draws the view.
func (v *PullsView) Layout(gtx giode.Context) giode.Dimensions {
	v.selectAll.SetChecked(v.sel.AllOf(v.keys()))
	head := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(v.slug, theme.CodeDim))}
	if v.linked {
		head = append([]giode.Element{v.selectAll}, append(head, components.Pill("server", v.account.Kind.Title(), theme.Info, theme.InfoSoft), v.reload)...)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 20}}, head...),
			components.ListFrame(v.body()),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 20, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, v.bulk),
		v.confirm,
	).Layout(gtx)
}
