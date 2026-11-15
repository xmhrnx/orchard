package components

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// NavItem is a sidebar destination with a glyph, a label, an optional
// count and a glowing marker while active.
type NavItem struct {
	press  *giode.PressableElement
	icon   string
	label  string
	active func() bool
	count  func() int
}

// NewNavItem creates a sidebar entry. Create it once.
func NewNavItem(icon, label string, active func() bool, count func() int, onClick func()) *NavItem {
	n := &NavItem{icon: icon, label: label, active: active, count: count}
	n.press = giode.Pressable(n.content).Label(label).OnClick(func(giode.ClickEvent) { onClick() })
	return n
}

func (n *NavItem) content(s giode.PressState) giode.Element {
	on := n.active()
	fg, bg, marker := theme.Muted, giode.Color("transparent"), giode.Color("transparent")
	switch {
	case on:
		fg, bg, marker = theme.Text, theme.Selected, theme.Primary
	case s.Hovered:
		fg, bg = theme.Text, theme.Hover
	}
	iconColor := fg
	if on {
		iconColor = theme.Primary
	}
	kids := []giode.Element{
		giode.Box(giode.Styles{Width: 3, Height: 18, BorderRadius: 3, Background: marker}),
		Glyph(n.icon, 17, iconColor),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(n.label, giode.Styles{FontSize: 13.5, FontWeight: giode.FontWeightMedium, Color: fg})),
	}
	if n.count != nil {
		if c := n.count(); c > 0 {
			kids = append(kids, giode.Text(strconv.Itoa(c), giode.Styles{FontSize: 11.5, Color: theme.Muted, FontFamily: theme.Mono}))
		}
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Top: 8, Bottom: 8, Right: 12}, BorderRadius: theme.Radius,
		Background: bg, Transition: theme.HoverFade,
	}, kids...)
}

// Layout draws the entry.
func (n *NavItem) Layout(gtx giode.Context) giode.Dimensions { return n.press.Layout(gtx) }
