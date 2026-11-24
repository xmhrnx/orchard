package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dropdown"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// RemoteBrowser lists an account's repositories and clones any number
// of them in one go, marking the ones already on this computer.
type RemoteBrowser struct {
	env       *env.Env
	jobs      *CloneJobs
	account   *dropdown.Select
	seen      string
	repos     []forge.RemoteRepo
	shown     []forge.RemoteRepo
	loading   bool
	sel       *components.Selection
	search    *components.SearchBox
	list      *giode.LazyListElement
	bulk      *components.BulkBar
	dest      *FolderField
	load      *giode.ButtonElement
	connect   *giode.ButtonElement
	loadedFor string
}

// NewRemoteBrowser creates the browser. Create it once.
func NewRemoteBrowser(e *env.Env, jobs *CloneJobs) *RemoteBrowser {
	b := &RemoteBrowser{env: e, jobs: jobs, sel: components.NewSelection()}
	b.account = giode.Input.Select(components.FieldStyle().With(giode.Styles{Width: 280}), "No accounts").OnChange(func(int) { b.fetch() })
	b.search = components.NewSearchBox("Filter repositories")
	b.dest = NewFolderField(e, "Clone into…")
	b.load = components.Button("Refresh", "fetch", components.Secondary, b.fetch)
	b.connect = components.Button("Connect an account", "user", components.Primary, func() { e.Nav.Open(env.PageAccounts) })
	b.list = giode.LazyList(
		func() int { return len(b.shown) },
		func(i int) any { return b.shown[i].FullName },
		func(i int) giode.Element { return b.row(b.shown[i]) },
		giode.Styles{Gap: 2, ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	b.bulk = components.NewBulkBar(b.sel, "repository", "repositories",
		components.Button("Clone selected", "clone", components.Primary, b.cloneSelected),
	)
	return b
}

// Prefill readies the browser when the page opens.
func (b *RemoteBrowser) Prefill() {
	if b.dest.Path() == "" {
		b.dest.SetPath(b.env.WS.Settings().CloneDir)
	}
	b.syncAccounts()
	if acc, ok := b.current(); ok && b.loadedFor != acc.ID {
		b.fetch()
	}
}

func (b *RemoteBrowser) syncAccounts() {
	var labels []string
	for _, a := range b.env.WS.Accounts() {
		labels = append(labels, a.Username+" @ "+a.Host)
	}
	key := strings.Join(labels, "\x00")
	if key == b.seen {
		return
	}
	b.seen = key
	if len(labels) == 0 {
		labels = []string{"No accounts"}
	}
	b.account.SetOptions(labels...)
}

func (b *RemoteBrowser) current() (forge.Account, bool) {
	accs := b.env.WS.Accounts()
	i := b.account.Selected()
	if i < 0 || i >= len(accs) {
		return forge.Account{}, false
	}
	return accs[i], true
}

func (b *RemoteBrowser) fetch() {
	acc, ok := b.current()
	if !ok {
		return
	}
	b.loading, b.loadedFor = true, acc.ID
	core.Query(b.env.WS, func(ctx context.Context) ([]forge.RemoteRepo, error) {
		client, err := b.env.WS.Client(acc)
		if err != nil {
			return nil, err
		}
		return client.Repositories(ctx)
	}, func(repos []forge.RemoteRepo, err error) {
		b.loading = false
		b.env.Fail("Listing repositories", err)
		b.repos = repos
		b.sel.Clear()
	})
}

// cloned reports whether a remote repository is already tracked.
func (b *RemoteBrowser) cloned(r forge.RemoteRepo) bool {
	for _, t := range b.env.WS.Repos() {
		if loc, ok := forge.ParseRemote(t.Origin); ok && strings.EqualFold(loc.Repo, r.FullName) {
			return true
		}
	}
	return false
}

func (b *RemoteBrowser) cloneSelected() {
	acc, _ := b.current()
	ssh := b.env.WS.Settings().PreferSSH
	n := 0
	for _, r := range b.repos {
		if !b.sel.Has(r.FullName) {
			continue
		}
		url := r.CloneHTTPS
		if ssh && r.CloneSSH != "" {
			url = r.CloneSSH
		}
		req := core.CloneRequest{URL: url, ParentDir: b.dest.Path(), Recursive: true}
		if !ssh {
			req.Account = &acc
		}
		b.jobs.Start(req)
		n++
	}
	b.sel.Clear()
	b.env.Toasts.Show(giode.ToastInfo, fmt.Sprintf("Cloning %d repositories…", n))
}

func (b *RemoteBrowser) row(r forge.RemoteRepo) giode.Element {
	return components.SelectRow("remote-repo", r.FullName, b.sel, func() giode.Element {
		badges := []giode.Element{giode.Text(r.FullName, theme.Ellipsis(theme.Strong))}
		if r.Private {
			badges = append(badges, components.Pill("lock", "private", theme.Muted, theme.Hover))
		}
		if r.Fork {
			badges = append(badges, components.Pill("branch", "fork", theme.Muted, theme.Hover))
		}
		if b.cloned(r) {
			badges = append(badges, components.Pill("check", "on this computer", theme.Primary, theme.PrimarySoft))
		}
		about := r.Description
		if about == "" {
			about = "No description"
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, badges...),
				giode.Text(about, theme.Ellipsis(theme.Caption))),
			giode.Row(giode.Styles{Gap: 4, Align: giode.AlignCenter},
				components.Glyph("star", 13, theme.Muted), giode.Text(fmt.Sprint(r.Stars), theme.Caption)),
			giode.Box(giode.Styles{Width: 80}, giode.Text(components.Ago(r.Updated), giode.Styles{FontSize: 12, Color: theme.Muted, TextAlign: giode.TextAlignEnd, Width: -1})),
		)
	})
}

func (b *RemoteBrowser) filter() {
	q := strings.ToLower(b.search.Query())
	b.shown = b.shown[:0]
	for _, r := range b.repos {
		if q == "" || strings.Contains(strings.ToLower(r.FullName+" "+r.Description), q) {
			b.shown = append(b.shown, r)
		}
	}
}

func (b *RemoteBrowser) body() giode.Element {
	switch {
	case len(b.env.WS.Accounts()) == 0:
		return components.EmptyState("user", "No accounts yet", "Connect GitHub, GitLab, Gitea/Forgejo or Bitbucket to browse and clone your repositories in bulk.", b.connect)
	case b.loading:
		return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(40), Justify: giode.JustifyCenter},
			giode.Spinner(giode.Styles{Width: 28, Height: 28, Color: theme.Primary}))
	case len(b.shown) == 0:
		return components.EmptyState("search", "No repositories", "Nothing matches on this account.", nil)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(8)}, b.list)
}

// Layout draws the browser.
func (b *RemoteBrowser) Layout(gtx giode.Context) giode.Dimensions {
	b.syncAccounts()
	b.filter()
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				b.account, giode.Box(giode.Styles{FlexGrow: 1}, b.search), b.load),
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				giode.Text("Clone into", theme.Caption), giode.Box(giode.Styles{FlexGrow: 1}, b.dest)),
			components.ListFrame(b.body()),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 20, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, b.bulk),
	).Layout(gtx)
}
