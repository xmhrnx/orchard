package components

import (
	"image"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/elements"

	"orchard/ui/theme"
)

// Popover is a panel that drops down under a trigger and stays open
// while it's used, unlike a menu, which closes on the first click.
// A click outside it or Escape closes it.
type Popover struct {
	open    bool
	content func() giode.Element
	outside *giode.PressableElement
	panel   giode.Element
}

// NewPopover creates a popover showing content. Create it once.
func NewPopover(content func() giode.Element) *Popover {
	p := &Popover{content: content}
	p.outside = giode.Pressable(func(giode.PressState) giode.Element {
		return giode.Box(giode.Styles{Width: -1, Height: -1})
	}).Label("Close").OnClick(func(giode.ClickEvent) { p.open = false })
	p.outside.Styles(giode.Styles{Cursor: giode.CursorDefault})
	p.panel = giode.Raw(func(gtx giode.Context) giode.Dimensions {
		// The portal hands over its whole slot as a minimum; the panel
		// keeps its own size within it.
		gtx.Constraints.Min = image.Point{}
		return giode.Box(giode.Styles{
			Padding: giode.UniformInset(14), BorderRadius: 14, Background: theme.Raised,
			BorderWidth: 1, BorderColor: giode.Mix("theme.primary", "theme.border", 0.25),
			BoxShadow: giode.BoxShadow{Y: 16, Blur: 40, Color: "#00000088"},
		}, p.content()).Layout(gtx)
	})
	return p
}

// Toggle opens or closes the popover.
func (p *Popover) Toggle() { p.open = !p.open }

// Close closes the popover.
func (p *Popover) Close() { p.open = false }

// Opened reports whether the popover is showing.
func (p *Popover) Opened() bool { return p.open }

// Anchor lays out trigger and, while open, the panel below it, its
// right edge lined up with the trigger's.
func (p *Popover) Anchor(trigger giode.Element) giode.Element {
	return giode.Raw(func(gtx giode.Context) giode.Dimensions {
		origin := elements.AbsolutePosition()
		dims := trigger.Layout(gtx)
		if !p.open {
			return dims
		}
		elements.HandleDismiss(p.Close)
		elements.RegisterOverlay(func() elements.Element { return p.outside })
		rect := image.Rectangle{Min: origin, Max: origin.Add(dims.Size)}
		elements.RegisterPortalAround(gtx, rect, gtx.Dp(6), elements.PlaceBelow, func() elements.Element {
			return giode.Raw(func(gtx giode.Context) giode.Dimensions {
				return p.panel.Layout(gtx)
			})
		})
		return dims
	})
}
