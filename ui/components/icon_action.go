package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/tooltip"

	"orchard/ui/theme"
)

// IconAction is a compact glyph button with a tooltip.
type IconAction struct {
	tip *tooltip.Tooltip
}

// NewIconAction creates an icon button. Create it once.
func NewIconAction(icon, hint string, onClick func()) *IconAction {
	btn := giode.IconButton(Glyph(icon, 16, theme.Muted), giode.Styles{
		Background:   theme.Text,
		BorderRadius: theme.RadiusSmall,
		Padding:      giode.UniformInset(6),
	}).Variant(giode.IconButtonGhost).Label(hint).OnClick(onClick)
	return &IconAction{tip: giode.Tooltip(btn, hint)}
}

// Layout draws the button.
func (a *IconAction) Layout(gtx giode.Context) giode.Dimensions { return a.tip.Layout(gtx) }
