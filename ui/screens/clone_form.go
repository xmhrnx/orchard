package screens

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/input"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// CloneForm clones any repository by URL, from any host.
type CloneForm struct {
	env       *env.Env
	jobs      *CloneJobs
	url       *input.Input
	dest      *FolderField
	folder    *input.Input
	branch    *input.Input
	group     *input.Input
	shallow   *checkbox.Checkbox
	recursive *checkbox.Checkbox
	clone     *giode.ButtonElement
}

// NewCloneForm creates the form. Create it once.
func NewCloneForm(e *env.Env, jobs *CloneJobs) *CloneForm {
	f := &CloneForm{env: e, jobs: jobs}
	f.url = giode.Input.Text("https://github.com/owner/repo.git  or  git@host:owner/repo.git", components.FieldStyle().With(giode.Styles{FontFamily: theme.Mono, FontSize: 13})).
		Keyboard(giode.KeyboardURL).Submit(func(string) { f.start() })
	f.dest = NewFolderField(e, "Clone into…")
	f.folder = giode.Input.Text("same as the repository", components.FieldStyle())
	f.branch = giode.Input.Text("default branch", components.FieldStyle())
	f.group = giode.Input.Text("none", components.FieldStyle())
	box := giode.Styles{Background: theme.Primary, BorderColor: theme.Muted, Color: theme.Text}
	f.shallow = giode.Input.Checkbox("Shallow (latest commit only)", box)
	f.recursive = giode.Input.Checkbox("Include submodules", box)
	f.recursive.SetChecked(true)
	f.clone = components.Button("Clone", "clone", components.Primary, f.start)
	return f
}

// Prefill sets the destination to the default clone folder.
func (f *CloneForm) Prefill() {
	if f.dest.Path() == "" {
		f.dest.SetPath(f.env.WS.Settings().CloneDir)
	}
}

func (f *CloneForm) start() {
	url := strings.TrimSpace(f.url.Text())
	if url == "" {
		f.env.Toasts.Show(giode.ToastWarning, "Paste the repository's URL first")
		f.url.Focus()
		return
	}
	req := core.CloneRequest{
		URL: url, ParentDir: f.dest.Path(), Folder: strings.TrimSpace(f.folder.Text()),
		Branch: strings.TrimSpace(f.branch.Text()), Shallow: f.shallow.Checked(),
		Recursive: f.recursive.Checked(), Group: strings.TrimSpace(f.group.Text()),
	}
	if loc, ok := forge.ParseRemote(url); ok && strings.HasPrefix(url, "http") {
		if acc, ok := forge.MatchAccount(f.env.WS.Accounts(), loc); ok {
			req.Account = &acc
		}
	}
	f.jobs.Start(req)
	f.url.SetText("")
	f.folder.SetText("")
}

// Layout draws the form.
func (f *CloneForm) Layout(gtx giode.Context) giode.Dimensions {
	target := "Will clone into " + components.ShortPath(f.dest.Path())
	if u := strings.TrimSpace(f.url.Text()); u != "" {
		req := core.CloneRequest{URL: u, ParentDir: f.dest.Path(), Folder: strings.TrimSpace(f.folder.Text())}
		target = "Will clone " + gitx.RepoNameFromURL(u) + " into " + components.ShortPath(req.Dest())
	}
	return components.Frame(giode.Styles{Width: -1, Gap: 18, Padding: giode.UniformInset(22)},
		components.Field("Repository URL", "HTTPS or SSH. Works with GitHub, GitLab, Gitea, Bitbucket, Azure DevOps, self-hosted servers: anything that speaks git.", f.url),
		components.Field("Destination", "", f.dest),
		giode.Row(giode.Styles{Width: -1, Gap: 14, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("Folder name", "", f.folder)),
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("Branch", "", f.branch)),
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("Group", "", f.group)),
		),
		giode.Row(giode.Styles{Width: -1, Gap: 22, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, f.shallow, f.recursive),
		giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(target, theme.Ellipsis(theme.CodeDim))), f.clone),
	).Layout(gtx)
}
