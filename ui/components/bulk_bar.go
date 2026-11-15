package components

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// BulkBar floats over a list while items are selected, offering the
// actions that apply to all of them at once. It slides up when the
// first item is checked and away when the selection empties.
type BulkBar struct {
	sel     *Selection
	noun    string
	actions []giode.Element
	clear   *IconAction
	anim    *giode.TransitionElement
}

// NewBulkBar creates the bar for a selection; noun names the items
// ("tags"). Create it once.
func NewBulkBar(sel *Selection, noun string, actions ...giode.Element) *BulkBar {
	b := &BulkBar{sel: sel, noun: noun, actions: actions}
	b.clear = NewIconAction("x", "Clear selection", sel.Clear)
	b.anim = giode.Transition(giode.Raw(b.layoutBar), giode.Fade|giode.SlideUp, theme.Smooth, giode.EaseOutBack).Hidden()
	return b
}

func (b *BulkBar) layoutBar(gtx giode.Context) giode.Dimensions {
	kids := []giode.Element{
		giode.Box(giode.Styles{Padding: giode.SymmetricInset(10, 4), BorderRadius: theme.RadiusPill, Background: theme.Primary},
			giode.Text(fmt.Sprint(b.sel.Count()), giode.Styles{FontSize: 12, FontWeight: giode.FontWeightBold, Color: theme.OnPrimary})),
		giode.Text(b.noun+" selected", giode.Styles{FontSize: 13, Color: theme.Text}),
		giode.Box(giode.Styles{Width: 1, Height: 22, Background: theme.Hairline, Margin: giode.SymmetricInset(4, 0)}),
	}
	kids = append(kids, b.actions...)
	kids = append(kids, b.clear)
	return Frame(giode.Styles{
		FlexDirection: giode.FlexDirectionRow, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Top: 8, Bottom: 8, Left: 10, Right: 8}, BorderRadius: 14,
		Background: theme.Raised,
	}, kids...).Layout(gtx)
}

// Layout draws the bar while anything is selected.
func (b *BulkBar) Layout(gtx giode.Context) giode.Dimensions {
	b.anim.SetVisible(b.sel.Count() > 0)
	return b.anim.Layout(gtx)
}
