package screens

import (
	"context"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// RemotesView shows where a repository syncs to and edits those links.
type RemotesView struct {
	env     *env.Env
	repo    *core.RepoState
	remotes []gitx.Remote
	list    *scroll.Scroll
	add     *giode.ButtonElement
	prompt  *components.PromptDialog
	confirm *components.ConfirmDialog
}

// NewRemotesView creates the view. Create it once.
func NewRemotesView(e *env.Env) *RemotesView {
	v := &RemotesView{env: e, prompt: components.NewPromptDialog(), confirm: components.NewConfirmDialog()}
	v.list = giode.Scroll(v.rows, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	v.add = components.Button("Add remote", "plus", components.Primary, v.addRemote)
	return v
}

// Open loads a repository's remotes.
func (v *RemotesView) Open(r *core.RepoState) {
	v.repo = r
	v.load()
}

func (v *RemotesView) load() {
	path := v.repo.Path
	core.Query(v.env.WS, func(ctx context.Context) ([]gitx.Remote, error) {
		return gitx.Remotes(ctx, path)
	}, func(list []gitx.Remote, err error) {
		v.env.Fail("Loading remotes", err)
		v.remotes = list
	})
}

func (v *RemotesView) act(busy, success string, fn func(ctx context.Context, path string) error) {
	v.env.WS.Do(v.repo.ID, busy, fn, func(err error) {
		v.env.Done(success, err)
		v.load()
	})
}

func (v *RemotesView) addRemote() {
	v.prompt.Ask("Add remote", "Name", "upstream", "upstream", "Next", func(name string) {
		v.prompt.Ask("Add remote "+name, "URL", "git@host:owner/repo.git", "", "Add", func(url string) {
			v.act("Adding remote", "Added "+name, func(ctx context.Context, path string) error { return gitx.AddRemote(ctx, path, name, url) })
		})
	})
}

// webURL turns a remote into the repository's web page, when it is on
// a recognisable host.
func webURL(remote string) string {
	loc, ok := forge.ParseRemote(remote)
	if !ok {
		return ""
	}
	return "https://" + loc.Host + "/" + loc.Repo
}

func (v *RemotesView) row(r gitx.Remote) giode.Element {
	urls := []giode.Element{giode.Text("fetch  "+r.FetchURL, theme.Ellipsis(theme.CodeDim))}
	if r.PushURL != "" && r.PushURL != r.FetchURL {
		urls = append(urls, giode.Text("push   "+r.PushURL, theme.Ellipsis(theme.CodeDim)))
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(14, 12), BorderRadius: theme.Radius, Background: theme.Raised,
	},
		giode.Box(giode.Styles{Padding: giode.UniformInset(8), BorderRadius: 9, Background: theme.InfoSoft}, components.Glyph("globe", 16, theme.Info)),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 4},
			giode.Text(r.Name, theme.Strong),
			giode.Stack(giode.Styles{Gap: 2}, urls...),
		),
		components.RowMenu("remote", r.Name,
			components.MenuItem{Icon: "external", Label: "Open in browser", Run: func() {
				if u := webURL(r.FetchURL); u != "" {
					v.env.OpenURL(u)
				}
			}},
			components.MenuItem{Icon: "copy", Label: "Copy URL", Run: func() { v.env.Copy("URL", r.FetchURL) }},
			components.MenuItem{Icon: "link", Label: "Change URL…", Run: func() {
				v.prompt.Ask("Change "+r.Name+" URL", "URL", "", r.FetchURL, "Save", func(url string) {
					v.act("Updating remote", "Remote updated", func(ctx context.Context, path string) error {
						return gitx.SetRemoteURL(ctx, path, r.Name, strings.TrimSpace(url))
					})
				})
			}},
			components.MenuSeparator,
			components.MenuItem{Icon: "trash", Label: "Remove", Run: func() {
				v.confirm.Ask("Remove "+r.Name+"?", "The remote and its tracking branches are removed from this repository. Nothing changes on the server.", "Remove", true, func() {
					v.act("Removing remote", "Removed "+r.Name, func(ctx context.Context, path string) error { return gitx.RemoveRemote(ctx, path, r.Name) })
				})
			}},
		),
	)
}

func (v *RemotesView) rows() giode.Element {
	rows := make([]giode.Element, len(v.remotes))
	for i, r := range v.remotes {
		rows[i] = v.row(r)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, rows...)
}

// Layout draws the view.
func (v *RemotesView) Layout(gtx giode.Context) giode.Dimensions {
	body := giode.Element(giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(12)}, v.list))
	if len(v.remotes) == 0 {
		body = components.EmptyState("globe", "No remotes", "This repository only lives here. Add a remote to back it up or share it.", nil)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd}, v.add),
		components.ListFrame(body),
		v.prompt, v.confirm,
	).Layout(gtx)
}
