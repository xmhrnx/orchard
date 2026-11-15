package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"orchard/internal/core"
	"orchard/ui/theme"
)

// RepoActions are what a repository card can ask for, by repository ID.
type RepoActions struct {
	Open      func(id string)
	Pull      func(id string)
	Fetch     func(id string)
	Push      func(id string)
	Reveal    func(id string)
	Terminal  func(id string)
	CopyPath  func(id string)
	TogglePin func(id string)
	Remove    func(id string)
}

// RepoCard is a repository in the overview: selectable, opens on click,
// with a one-click update and a menu of everything else.
type RepoCard struct {
	repo  *core.RepoState
	sel   *Selection
	box   *checkbox.Checkbox
	press *giode.PressableElement
	pull  *IconAction
	more  giode.Element
	menu  *giode.MenuElement
}

// RepoCardFor returns the card of a repository, keeping its state
// across frames.
func RepoCardFor(r *core.RepoState, sel *Selection, act *RepoActions) *RepoCard {
	c := giode.Keyed("repo-card", r.ID, func() *RepoCard { return newRepoCard(r.ID, sel, act) })
	c.repo = r
	return c
}

func newRepoCard(id string, sel *Selection, act *RepoActions) *RepoCard {
	c := &RepoCard{sel: sel}
	c.box = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
		OnChange(func(on bool) { sel.Set(id, on) })
	c.pull = NewIconAction("pull", "Pull", func() { act.Pull(id) })
	c.menu = giode.Menu().
		IconItem(GlyphSource("pull"), "Pull", func() { act.Pull(id) }).
		IconItem(GlyphSource("fetch"), "Fetch", func() { act.Fetch(id) }).
		IconItem(GlyphSource("push"), "Push", func() { act.Push(id) }).
		Separator().
		IconItem(GlyphSource("folder"), "Show in file manager", func() { act.Reveal(id) }).
		IconItem(GlyphSource("terminal"), "Open terminal here", func() { act.Terminal(id) }).
		IconItem(GlyphSource("copy"), "Copy path", func() { act.CopyPath(id) }).
		IconItem(GlyphSource("pin"), "Pin / unpin", func() { act.TogglePin(id) }).
		Separator().
		IconItem(GlyphSource("trash"), "Stop tracking", func() { act.Remove(id) })
	c.more = c.menu.Trigger(NewIconAction("dots", "More", func() {}))
	c.press = giode.Pressable(c.content).Label("Open " + id).OnClick(func(giode.ClickEvent) { act.Open(id) })
	c.press.Context(c.menu)
	return c
}

func (c *RepoCard) content(s giode.PressState) giode.Element {
	r := c.repo
	border, bg := theme.Hairline, theme.Surface
	if c.sel.Has(r.ID) {
		border, bg = giode.Mix("theme.primary", "theme.border", 0.55), theme.Selected
	} else if s.Hovered {
		border = giode.Mix("theme.primary", "theme.border", 0.3)
	}
	subtitle := giode.Text(ShortPath(r.Path), theme.Ellipsis(theme.CodeDim))
	if r.Err != nil && r.Busy == "" {
		subtitle = giode.Text(r.Err.Error(), theme.Ellipsis(theme.Tinted(theme.Caption, theme.Danger)))
	}
	name := []giode.Element{giode.Text(r.Name, theme.Ellipsis(giode.Styles{FontSize: 14.5, FontWeight: giode.FontWeightSemiBold, Color: theme.Text}))}
	if r.Pinned {
		name = append(name, Glyph("pin", 12, theme.Accent))
	}
	if r.Group != "" {
		name = append(name, Pill("", r.Group, theme.Muted, theme.Hover))
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 14, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Top: 12, Bottom: 12, Left: 14, Right: 10}, BorderRadius: 14,
		Background: bg, BorderWidth: 1, BorderColor: border, Transition: theme.HoverFade,
	},
		c.box,
		giode.Box(giode.Styles{Padding: giode.UniformInset(9), BorderRadius: 11, Background: theme.PrimarySoft}, Glyph("repo", 18, theme.Primary)),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 4, MinWidth: 120},
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, name...),
			subtitle,
		),
		BranchLabel(r.Status),
		giode.Box(giode.Styles{MinWidth: 150}, RepoStatus(r)),
		c.pull,
		c.more,
	)
}

// Layout draws the card.
func (c *RepoCard) Layout(gtx giode.Context) giode.Dimensions {
	c.box.SetChecked(c.sel.Has(c.repo.ID))
	return giode.Box(giode.Styles{Width: -1}, c.press, c.menu).Layout(gtx)
}
