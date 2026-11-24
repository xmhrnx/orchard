package screens

import (
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"

	"orchard/internal/core"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// RemoveReposDialog removes repositories: from Orchard's list only, or
// with their folders, which go to the system trash. Before deleting it
// names any work that exists nowhere else.
type RemoveReposDialog struct {
	env     *env.Env
	open    bool
	ids     []string
	removed func(ids []string)
	cancel  *giode.ButtonElement
	untrack *giode.ButtonElement
	trash   *giode.ButtonElement
	dlg     *dialog.Dialog
}

// NewRemoveReposDialog creates the dialog; removed runs with the IDs
// that left the list. Create it once.
func NewRemoveReposDialog(e *env.Env, removed func(ids []string)) *RemoveReposDialog {
	d := &RemoveReposDialog{env: e, removed: removed}
	d.cancel = components.Button("Cancel", "", components.Secondary, func() { d.open = false })
	d.untrack = components.Button("Remove from list", "minus", components.Secondary, d.removeFromList)
	d.trash = components.Button("Move to trash", "trash", components.Destructive, d.moveToTrash)
	d.dlg = giode.Dialog(&d.open, d.content, components.DialogPanel())
	return d
}

// Ask opens the dialog for repositories.
func (d *RemoveReposDialog) Ask(ids []string) {
	if len(ids) == 0 {
		return
	}
	d.ids, d.open = append([]string(nil), ids...), true
}

func (d *RemoveReposDialog) repos() []*core.RepoState {
	var out []*core.RepoState
	for _, id := range d.ids {
		if r := d.env.WS.Repo(id); r != nil {
			out = append(out, r)
		}
	}
	return out
}

func (d *RemoveReposDialog) removeFromList() {
	ids := d.ids
	d.open = false
	giode.AfterLayout(func() {
		if err := d.env.WS.Remove(ids...); err != nil {
			d.env.Fail("Removing", err)
			return
		}
		d.env.Toasts.Show(giode.ToastSuccess, "Removed "+count(len(ids), "repository", "repositories")+" from the list")
		d.removed(ids)
	})
}

func (d *RemoveReposDialog) moveToTrash() {
	ids := d.ids
	d.open = false
	d.env.WS.Trash(ids, func(trashed int, err error) {
		if trashed > 0 {
			d.env.Toasts.Show(giode.ToastSuccess, "Moved "+count(trashed, "repository", "repositories")+" to the trash")
			d.removed(ids)
		}
		d.env.Fail("Deleting", err)
	})
}

// atRisk lists the work in these repositories that deleting would lose.
func atRisk(repos []*core.RepoState) []string {
	var lines []string
	for _, r := range repos {
		st := r.Status
		var what []string
		if n := st.ChangeCount(); n > 0 {
			what = append(what, count(n, "uncommitted change", "uncommitted changes"))
		}
		if st.Ahead > 0 {
			what = append(what, count(st.Ahead, "unpushed commit", "unpushed commits"))
		}
		if st.Upstream == "" && !st.Detached && r.Origin != "" {
			what = append(what, "a branch never pushed")
		}
		if r.Origin == "" {
			what = append(what, "no remote copy")
		}
		if st.Stashes > 0 {
			what = append(what, count(st.Stashes, "stash", "stashes"))
		}
		if len(what) > 0 {
			lines = append(lines, r.Name+": "+strings.Join(what, ", "))
		}
	}
	return lines
}

func (d *RemoveReposDialog) content() giode.Element {
	repos := d.repos()
	title := fmt.Sprintf("Remove %s?", count(len(repos), "repository", "repositories"))
	if len(repos) == 1 {
		title = "Remove " + repos[0].Name + "?"
	}
	kids := []giode.Element{
		giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Padding: giode.UniformInset(9), BorderRadius: 12, Background: theme.DangerSoft}, components.Glyph("trash", 20, theme.Danger)),
			giode.Text(title, theme.Heading)),
		giode.Text("Remove from list keeps the files where they are. Move to trash deletes the folder too; you can restore it from the Trash or Recycle Bin.",
			giode.Styles{FontSize: 13, Color: theme.Muted, LineHeight: 19}),
	}
	if risk := atRisk(repos); len(risk) > 0 {
		lines := []giode.Element{giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			components.Glyph("alert", 14, theme.Warning),
			giode.Text("Work that exists only on this computer", giode.Styles{FontSize: 12.5, FontWeight: giode.FontWeightSemiBold, Color: theme.Warning}))}
		for _, l := range risk {
			lines = append(lines, giode.Text(l, giode.Styles{FontSize: 12.5, Color: theme.Text}))
		}
		kids = append(kids, giode.Stack(giode.Styles{Width: -1, Gap: 4, Padding: giode.UniformInset(12), BorderRadius: theme.Radius, Background: theme.AccentSoft}, lines...))
	}
	kids = append(kids, giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, d.untrack, d.trash))
	return giode.Stack(giode.Styles{Gap: 14, Width: 500}, kids...)
}

// Layout draws the dialog.
func (d *RemoveReposDialog) Layout(gtx giode.Context) giode.Dimensions { return d.dlg.Layout(gtx) }
