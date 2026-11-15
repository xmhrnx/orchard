package screens

import (
	"context"
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// fileSel identifies the file whose diff is shown.
type fileSel struct {
	change gitx.FileChange
	staged bool
}

// ChangesView is the work in progress: staged and unstaged files, the
// diff of the one selected, and the commit box.
type ChangesView struct {
	env        *env.Env
	repoID     string
	selected   *fileSel
	diff       *components.DiffView
	files      *scroll.Scroll
	message    *input.Input
	amend      *checkbox.Checkbox
	commit     *giode.ButtonElement
	commitPush *giode.ButtonElement
	stageAll   *components.IconAction
	unstageAll *components.IconAction
	discardAll *components.IconAction
	stageOne   *giode.ButtonElement
	unstageOne *giode.ButtonElement
	discardOne *giode.ButtonElement
	confirm    *components.ConfirmDialog
}

// NewChangesView creates the view. Create it once.
func NewChangesView(e *env.Env) *ChangesView {
	v := &ChangesView{env: e, diff: components.NewDiffView(), confirm: components.NewConfirmDialog()}
	v.files = giode.Scroll(v.fileLists, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	v.message = giode.Input.Text("Describe what changed…", components.FieldStyle()).Lines(3, 8)
	v.amend = giode.Input.Checkbox("Amend last commit", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted, Color: theme.Muted})
	v.commit = components.Button("Commit", "check", components.Primary, func() { v.doCommit(false) })
	v.commitPush = components.Button("Commit & push", "push", components.Secondary, func() { v.doCommit(true) })
	v.stageAll = components.NewIconAction("plus", "Stage everything", func() { v.stage(nil) })
	v.unstageAll = components.NewIconAction("minus", "Unstage everything", func() { v.unstage(nil) })
	v.discardAll = components.NewIconAction("undo", "Discard all changes", v.confirmDiscardAll)
	v.stageOne = components.Button("Stage", "plus", components.Secondary, func() { v.stage([]string{v.selected.change.Path}) })
	v.unstageOne = components.Button("Unstage", "minus", components.Secondary, func() { v.unstage([]string{v.selected.change.Path}) })
	v.discardOne = components.Button("Discard", "undo", components.Destructive, v.confirmDiscardOne)
	return v
}

// Open shows a repository's changes.
func (v *ChangesView) Open(r *core.RepoState) {
	if v.repoID != r.ID {
		v.selected = nil
		v.diff.SetLines(nil)
		v.message.SetText("")
	}
	v.repoID = r.ID
	v.env.WS.Refresh(r.ID)
}

func (v *ChangesView) repo() *core.RepoState {
	if r := v.env.WS.Repo(v.repoID); r != nil {
		return r
	}
	return &core.RepoState{}
}

func (v *ChangesView) show(c gitx.FileChange, staged bool) {
	v.selected = &fileSel{change: c, staged: staged}
	path := v.repo().Path
	core.Query(v.env.WS, func(ctx context.Context) ([]gitx.DiffLine, error) {
		return gitx.FileDiff(ctx, path, c, staged)
	}, func(lines []gitx.DiffLine, err error) {
		if err != nil {
			v.env.Fail("Loading diff", err)
			return
		}
		if v.selected != nil && v.selected.change.Path == c.Path && v.selected.staged == staged {
			v.diff.SetLines(lines)
		}
	})
}

// after refreshes the open diff once an index change lands.
func (v *ChangesView) after(what string) func(error) {
	return func(err error) {
		v.env.Fail(what, err)
		v.selected = nil
		v.diff.SetLines(nil)
	}
}

func (v *ChangesView) stage(paths []string) {
	v.env.WS.Do(v.repoID, "Staging", func(ctx context.Context, path string) error {
		return gitx.Stage(ctx, path, paths...)
	}, v.after("Staging"))
}

func (v *ChangesView) unstage(paths []string) {
	v.env.WS.Do(v.repoID, "Unstaging", func(ctx context.Context, path string) error {
		return gitx.Unstage(ctx, path, paths...)
	}, v.after("Unstaging"))
}

func (v *ChangesView) discard(tracked, untracked []string) {
	v.env.WS.Do(v.repoID, "Discarding", func(ctx context.Context, path string) error {
		return gitx.Discard(ctx, path, tracked, untracked)
	}, v.after("Discarding"))
}

func (v *ChangesView) confirmDiscardOne() {
	c := v.selected.change
	v.confirm.Ask("Discard changes to "+c.Path+"?", "Your edits to this file will be lost. This can't be undone.", "Discard", true, func() {
		if c.Kind == gitx.Untracked {
			v.discard(nil, []string{c.Path})
		} else {
			v.discard([]string{c.Path}, nil)
		}
	})
}

func (v *ChangesView) confirmDiscardAll() {
	st := v.repo().Status
	var tracked, untracked []string
	for _, c := range st.Unstaged {
		tracked = append(tracked, c.Path)
	}
	for _, c := range st.Untracked {
		untracked = append(untracked, c.Path)
	}
	body := fmt.Sprintf("%d modified and %d new files will be reset. Staged changes are kept. This can't be undone.", len(tracked), len(untracked))
	v.confirm.Ask("Discard all unstaged changes?", body, "Discard all", true, func() { v.discard(tracked, untracked) })
}

func (v *ChangesView) doCommit(push bool) {
	msg := v.message.Text()
	amend := v.amend.Checked()
	if msg == "" && !amend {
		v.env.Toasts.Show(giode.ToastWarning, "Write a commit message first")
		v.message.Focus()
		return
	}
	stageAll := len(v.repo().Status.Staged) == 0
	v.env.WS.Do(v.repoID, "Committing", func(ctx context.Context, path string) error {
		if stageAll {
			if err := gitx.Stage(ctx, path); err != nil {
				return err
			}
		}
		opts := gitx.CommitOptions{Message: msg, Amend: amend}
		if amend && msg == "" {
			body, err := gitx.CommitBody(ctx, path, "HEAD")
			if err != nil {
				return err
			}
			opts.Message = body
		}
		if err := gitx.CreateCommit(ctx, path, opts); err != nil {
			return err
		}
		if push {
			return gitx.Push(ctx, path, amend)
		}
		return nil
	}, func(err error) {
		success := "Committed"
		if push {
			success = "Committed and pushed"
		}
		v.env.Done(success, err)
		if err == nil {
			v.message.SetText("")
			v.amend.SetChecked(false)
			v.selected = nil
			v.diff.SetLines(nil)
		}
	})
}

func (v *ChangesView) fileRow(c gitx.FileChange, staged bool) giode.Element {
	key := fmt.Sprintf("%v:%s", staged, c.Path)
	return giode.Keyed("change-row", key, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			on := v.selected != nil && v.selected.staged == staged && v.selected.change.Path == c.Path
			return components.ChangeRow(c, on, s.Hovered)
		}).Label(c.Path).OnClick(func(giode.ClickEvent) { v.show(c, staged) }).
			OnDoubleClick(func(giode.ClickEvent) {
				if staged {
					v.unstage([]string{c.Path})
				} else {
					v.stage([]string{c.Path})
				}
			})
	})
}

func (v *ChangesView) group(title string, count int, actions []giode.Element, rows []giode.Element) giode.Element {
	head := append([]giode.Element{
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(fmt.Sprintf("%s · %d", title, count), theme.Overline)),
	}, actions...)
	kids := []giode.Element{giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Padding: giode.Inset{Left: 10, Right: 4}, Wrap: giode.WrapNoWrap}, head...)}
	if len(rows) == 0 {
		kids = append(kids, giode.Box(giode.Styles{Padding: giode.SymmetricInset(10, 6)}, giode.Text("Nothing here", giode.Styles{FontSize: 12, Color: theme.Faint})))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2}, append(kids, rows...)...)
}

func (v *ChangesView) fileLists() giode.Element {
	st := v.repo().Status
	var staged, unstaged []giode.Element
	for _, c := range st.Staged {
		staged = append(staged, v.fileRow(c, true))
	}
	for _, list := range [][]gitx.FileChange{st.Conflicted, st.Unstaged, st.Untracked} {
		for _, c := range list {
			unstaged = append(unstaged, v.fileRow(c, false))
		}
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 16},
		v.group("Staged", len(staged), []giode.Element{v.unstageAll}, staged),
		v.group("Changes", len(unstaged), []giode.Element{v.discardAll, v.stageAll}, unstaged),
	)
}

func (v *ChangesView) commitBox() giode.Element {
	st := v.repo().Status
	hint := fmt.Sprintf("%d staged", len(st.Staged))
	if len(st.Staged) == 0 {
		hint = "nothing staged: commit will include every change"
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.Inset{Top: 12}},
		v.message,
		giode.Text(hint, theme.Caption),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, v.amend), v.commitPush, v.commit),
	)
}

func (v *ChangesView) diffPane() giode.Element {
	if v.selected == nil {
		return components.EmptyState("eye", "Pick a file", "Select a changed file to see exactly what changed. Double-click to stage or unstage it.", nil)
	}
	c := v.selected.change
	buttons := []giode.Element{v.stageOne, v.discardOne}
	if v.selected.staged {
		buttons = []giode.Element{v.unstageOne}
	}
	head := append([]giode.Element{
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(c.Path, theme.Ellipsis(theme.Code))),
	}, buttons...)
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, head...),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1, BorderRadius: theme.Radius, Background: theme.Sunken, Overflow: giode.OverflowHidden}, v.diff),
	)
}

// Layout draws the view.
func (v *ChangesView) Layout(gtx giode.Context) giode.Dimensions {
	busy := v.repo().Busy != ""
	v.commit.Loading(v.repo().Busy == "Committing")
	v.commitPush.Disabled(busy)
	return giode.Row(giode.Styles{Width: -1, Height: -1, Gap: 14, Wrap: giode.WrapNoWrap},
		components.Frame(giode.Styles{Width: 400, Height: -1, Padding: giode.UniformInset(14)},
			giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, v.files),
			giode.Divider(giode.Styles{Color: theme.Hairline}),
			v.commitBox(),
		),
		components.FilledFrame(giode.Styles{FlexGrow: 1, Height: -1, Padding: giode.UniformInset(16)}, v.diffPane()),
		v.confirm,
	).Layout(gtx)
}
