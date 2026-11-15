package components

import (
	"strconv"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/ui/theme"
)

// Tab is one entry of a TabStrip. Count, when set, shows a number.
type Tab struct {
	Label string
	Icon  string
	Count func() int
}

// TabStrip is a row of tabs in a sunken track; the active tab rises out
// of it as a raised chip with an underline. When the tabs don't fit,
// the track scrolls sideways instead of wrapping.
type TabStrip struct {
	tabs     []Tab
	active   int
	items    []*giode.PressableElement
	onChange func(int)
	track    *scroll.Scroll
}

// NewTabStrip creates the tabs. Create it once.
func NewTabStrip(onChange func(int), tabs ...Tab) *TabStrip {
	t := &TabStrip{tabs: tabs, onChange: onChange}
	t.track = giode.Scroll(t.row, giode.Styles{ScrollBar: giode.ScrollBar{Hidden: true}}).Axis(giode.Horizontal)
	for i := range tabs {
		t.items = append(t.items, giode.Pressable(func(s giode.PressState) giode.Element {
			return t.tab(i, s)
		}).Label(tabs[i].Label).OnClick(func(giode.ClickEvent) { t.Select(i) }))
	}
	return t
}

// Active is the selected tab index.
func (t *TabStrip) Active() int { return t.active }

// Select activates tab i.
func (t *TabStrip) Select(i int) {
	if i == t.active || i < 0 || i >= len(t.tabs) {
		return
	}
	t.active = i
	if t.onChange != nil {
		t.onChange(i)
	}
}

func (t *TabStrip) tab(i int, s giode.PressState) giode.Element {
	tab := t.tabs[i]
	on := i == t.active
	fg, bg := theme.Muted, giode.Color("transparent")
	switch {
	case on:
		fg, bg = theme.Text, theme.Surface
	case s.Hovered:
		fg = theme.Text
	}
	kids := []giode.Element{Glyph(tab.Icon, 15, fg), giode.Text(tab.Label, giode.Styles{FontSize: 13, FontWeight: giode.FontWeightMedium, Color: fg})}
	if tab.Count != nil {
		if n := tab.Count(); n > 0 {
			kids = append(kids, Pill("", strconv.Itoa(n), theme.Primary, theme.PrimarySoft))
		}
	}
	underline := giode.Color("transparent")
	if on {
		underline = theme.Primary
	}
	st := giode.Styles{Padding: giode.SymmetricInset(10, 7), BorderRadius: 8, Background: bg, Transition: theme.HoverFade}
	if on {
		st.BoxShadow = giode.BoxShadow{Y: 2, Blur: 8, Color: "#00000040"}
	}
	return giode.Stack(giode.Styles{Gap: 3, Align: giode.AlignCenter},
		giode.Row(giode.Styles{Gap: 7, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}.With(st), kids...),
		giode.Box(giode.Styles{Width: 22, Height: 2, BorderRadius: 2, Background: underline}),
	)
}

func (t *TabStrip) row() giode.Element {
	els := make([]giode.Element, len(t.items))
	for i, it := range t.items {
		els[i] = it
	}
	return giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Top: 4, Left: 4, Right: 4}}, els...)
}

// Layout draws the strip.
func (t *TabStrip) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Box(giode.Styles{
		BorderRadius: 12, Background: theme.Sunken, Overflow: giode.OverflowHidden,
		BorderWidth: 1, BorderColor: theme.Hairline,
	}, t.track).Layout(gtx)
}
