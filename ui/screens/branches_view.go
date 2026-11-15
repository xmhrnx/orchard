package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/segmented"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// BranchesView lists local and remote branches for switching, merging
// and cleaning up several at once.
type BranchesView struct {
	env      *env.Env
	repo     *core.RepoState
	branches []gitx.Branch
	shown    []gitx.Branch
	sel      *components.Selection
	search   *components.SearchBox
	scope    *segmented.Control
	list     *giode.LazyListElement
	bulk     *components.BulkBar
	create   *giode.ButtonElement
	prune    *giode.ButtonElement
	reload   *components.IconAction
	prompt   *components.PromptDialog
	confirm  *components.ConfirmDialog
}

// NewBranchesView creates the view. Create it once.
func NewBranchesView(e *env.Env) *BranchesView {
	v := &BranchesView{env: e, sel: components.NewSelection()}
	v.search = components.NewSearchBox("Filter branches")
	v.scope = giode.Input.Segmented("Local", "Remote", "All")
	v.prompt = components.NewPromptDialog()
	v.confirm = components.NewConfirmDialog()
	v.reload = components.NewIconAction("fetch", "Reload", v.load)
	v.create = components.Button("New branch", "plus", components.Primary, v.newBranch)
	v.prune = components.Button("Prune merged", "eraser", components.Secondary, v.pruneMerged)
	v.list = giode.LazyList(
		func() int { return len(v.shown) },
		func(i int) any { return branchKey(v.shown[i]) },
		func(i int) giode.Element { return v.row(v.shown[i]) },
		giode.Styles{Gap: 2, ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	v.bulk = components.NewBulkBar(v.sel, "branches",
		components.Button("Delete", "trash", components.Destructive, func() { v.confirmDelete(false) }),
		components.Button("Force delete", "alert", components.Destructive, func() { v.confirmDelete(true) }),
	)
	return v
}

func branchKey(b gitx.Branch) string {
	if b.Remote {
		return "remote:" + b.Name
	}
	return "local:" + b.Name
}

// Open loads a repository's branches.
func (v *BranchesView) Open(r *core.RepoState) {
	if v.repo == nil || v.repo.ID != r.ID {
		v.sel.Clear()
	}
	v.repo = r
	v.load()
}

func (v *BranchesView) load() {
	path := v.repo.Path
	core.Query(v.env.WS, func(ctx context.Context) ([]gitx.Branch, error) {
		return gitx.Branches(ctx, path)
	}, func(list []gitx.Branch, err error) {
		v.env.Fail("Loading branches", err)
		v.branches = list
		keys := make([]string, len(list))
		for i, b := range list {
			keys[i] = branchKey(b)
		}
		v.sel.Retain(keys)
	})
}

// act runs a git action and reloads the list.
func (v *BranchesView) act(busy, success string, fn func(ctx context.Context, path string) error) {
	v.env.WS.Do(v.repo.ID, busy, fn, func(err error) {
		v.env.Done(success, err)
		v.load()
	})
}

func (v *BranchesView) newBranch() {
	v.prompt.Ask("New branch", "Branch name (created from the current commit)", "feature/…", "", "Create & switch", func(name string) {
		v.act("Branching", "Switched to "+name, func(ctx context.Context, path string) error {
			return gitx.CreateBranch(ctx, path, name, "", true)
		})
	})
}

func (v *BranchesView) pruneMerged() {
	v.env.WS.Do(v.repo.ID, "Pruning", func(ctx context.Context, path string) error {
		gone, err := gitx.PruneMergedBranches(ctx, path)
		v.env.App.Post(func() {
			if err == nil {
				v.env.Toasts.Show(giode.ToastSuccess, fmt.Sprintf("Removed %d merged branches", len(gone)))
			}
		})
		return err
	}, func(err error) {
		v.env.Fail("Pruning", err)
		v.load()
	})
}

func (v *BranchesView) confirmDelete(force bool) {
	var targets []gitx.Branch
	remote := 0
	for _, b := range v.branches {
		if v.sel.Has(branchKey(b)) {
			targets = append(targets, b)
			if b.Remote {
				remote++
			}
		}
	}
	body := fmt.Sprintf("%d branches will be deleted.", len(targets))
	if remote > 0 {
		body += fmt.Sprintf(" %d of them are remote: they will be deleted on the server for everyone.", remote)
	}
	if force {
		body += " Unmerged work on local branches will be lost."
	}
	v.confirm.Ask("Delete branches?", body, "Delete", true, func() {
		v.sel.Clear()
		v.act("Deleting branches", fmt.Sprintf("Deleted %d branches", len(targets)), func(ctx context.Context, path string) error {
			return gitx.DeleteBranches(ctx, path, targets, force)
		})
	})
}

func (v *BranchesView) menu(b gitx.Branch) giode.Element {
	return components.RowMenu("branch", branchKey(b),
		components.MenuItem{Icon: "check", Label: "Check out", Run: func() {
			v.act("Switching", "Switched to "+b.Name, func(ctx context.Context, path string) error { return gitx.Checkout(ctx, path, b) })
		}},
		components.MenuItem{Icon: "merge", Label: "Merge into current branch", Run: func() {
			v.act("Merging", "Merged "+b.Name, func(ctx context.Context, path string) error { return gitx.Merge(ctx, path, b.Name, false) })
		}},
		components.MenuItem{Icon: "branch", Label: "Rebase current branch onto this", Run: func() {
			v.act("Rebasing", "Rebased onto "+b.Name, func(ctx context.Context, path string) error { return gitx.Rebase(ctx, path, b.Name) })
		}},
		components.MenuSeparator,
		components.MenuItem{Icon: "settings", Label: "Rename…", Run: func() { v.rename(b) }},
		components.MenuItem{Icon: "copy", Label: "Copy name", Run: func() { v.env.Copy("Branch name", b.Name) }},
	)
}

func (v *BranchesView) rename(b gitx.Branch) {
	if b.Remote {
		v.env.Toasts.Show(giode.ToastInfo, "Remote branches are renamed on the server")
		return
	}
	v.prompt.Ask("Rename "+b.Name, "New name", "", b.Name, "Rename", func(name string) {
		v.act("Renaming", "Renamed to "+name, func(ctx context.Context, path string) error { return gitx.RenameBranch(ctx, path, b.Name, name) })
	})
}

func (v *BranchesView) row(b gitx.Branch) giode.Element {
	return components.SelectRow("branch-row", branchKey(b), v.sel, func() giode.Element {
		name := []giode.Element{}
		if b.Current {
			name = append(name, components.Pill("check", "current", theme.OnPrimary, theme.Primary))
		}
		icon, tint := "branch", theme.Primary
		if b.Remote {
			icon, tint = "globe", theme.Info
		}
		name = append(name, components.Glyph(icon, 14, tint), giode.Text(b.Name, theme.Ellipsis(theme.Tinted(theme.Code, theme.Text))))
		if b.Track != "" {
			name = append(name, components.Pill("", strings.Trim(b.Track, "[]"), theme.Accent, theme.AccentSoft))
		} else if b.Upstream != "" {
			name = append(name, components.Pill("link", b.Upstream, theme.Muted, theme.Hover))
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Width: 320}, name...),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(b.Subject, theme.Ellipsis(theme.Caption))),
			giode.Text(b.Hash, theme.CodeDim),
			v.menu(b),
		)
	})
}

func (v *BranchesView) filter() {
	q := strings.ToLower(v.search.Query())
	scope := v.scope.Selected()
	v.shown = v.shown[:0]
	for _, b := range v.branches {
		if (scope == 0 && b.Remote) || (scope == 1 && !b.Remote) {
			continue
		}
		if q == "" || strings.Contains(strings.ToLower(b.Name), q) {
			v.shown = append(v.shown, b)
		}
	}
}

// Layout draws the view.
func (v *BranchesView) Layout(gtx giode.Context) giode.Dimensions {
	v.filter()
	list := giode.Element(giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(8)}, v.list))
	if len(v.shown) == 0 {
		list = components.EmptyState("branch", "No branches here", "Try another filter.", nil)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				giode.Box(giode.Styles{Width: 300}, v.search), v.scope,
				giode.Box(giode.Styles{FlexGrow: 1}), v.reload, v.prune, v.create),
			components.ListFrame(list),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 20, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, v.bulk),
		v.prompt, v.confirm,
	).Layout(gtx)
}
