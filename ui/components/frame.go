package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"orchard/ui/theme"
)

// Frame is Orchard's signature panel: a surface inside a one-dp rim
// that catches light at the top and fades towards the bottom, with a
// soft drop shadow. st styles the inner surface (padding, gap, layout).
func Frame(st giode.Styles, kids ...giode.Element) giode.Element {
	return frameBox(st, false, kids)
}

// frameBox builds a frame; fillInner makes the surface fill the frame
// even when the content is shorter, for frames sized by their parent.
func frameBox(st giode.Styles, fillInner bool, kids []giode.Element) giode.Element {
	radius := st.BorderRadius
	if radius == 0 {
		radius = theme.RadiusLarge
	}
	inner := st
	inner.BorderRadius = radius - 1
	if inner.Background == "" {
		inner.Background = theme.Surface
	}
	if inner.FlexDirection == "" {
		inner.FlexDirection = giode.FlexDirectionColumn
	}
	inner.Width, inner.Height, inner.FlexGrow = properties.Width(fill(int(st.Width))), properties.Height(fill(int(st.Height))), 0
	if fillInner {
		inner.Width, inner.Height = -1, -1
	}
	inner.Margin = giode.Inset{}
	return giode.Box(giode.Styles{
		Width:              st.Width,
		Height:             st.Height,
		FlexGrow:           st.FlexGrow,
		Margin:             st.Margin,
		Padding:            giode.UniformInset(1),
		BorderRadius:       radius,
		BackgroundGradient: frameRim,
		BoxShadow:          giode.BoxShadow{Y: 10, Blur: 30, Color: theme.Shadow},
	}, giode.Box(inner, kids...))
}

// frameRim is the gradient that forms the rim.
var frameRim = giode.Linear(180,
	giode.Mix("theme.primary", "theme.border", 0.28),
	"theme.border",
	giode.Mix("theme.border", "theme.background", 0.5),
)

// fill makes the inner surface fill a frame with a fixed or filling size.
func fill(v int) int {
	if v != 0 {
		return -1
	}
	return 0
}
