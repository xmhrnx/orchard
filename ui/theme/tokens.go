package theme

import (
	"time"

	"github.com/xmhrnx/giode"
)

// Colors derived from the palette. Mixing theme tokens yields
// color-mix values that follow light/dark switches.
var (
	Canvas      giode.Color = "theme.background"
	Surface     giode.Color = "theme.surface"
	Raised                  = giode.Mix("theme.text", "theme.surface", 0.035)
	Sunken                  = giode.Mix("theme.background", "theme.surface", 0.5)
	Hover                   = giode.Mix("theme.text", "theme.surface", 0.06)
	Pressed                 = giode.Mix("theme.text", "theme.surface", 0.10)
	Hairline    giode.Color = "theme.border"
	Text        giode.Color = "theme.text"
	Muted       giode.Color = "theme.muted"
	Faint                   = giode.Mix("theme.muted", "theme.surface", 0.6)
	Primary     giode.Color = "theme.primary"
	OnPrimary   giode.Color = "theme.onprimary"
	Accent      giode.Color = "theme.accent"
	Danger      giode.Color = "theme.danger"
	Warning     giode.Color = "theme.warning"
	Info        giode.Color = "theme.info"
	PrimarySoft             = giode.Mix("theme.primary", "theme.surface", 0.14)
	AccentSoft              = giode.Mix("theme.accent", "theme.surface", 0.16)
	DangerSoft              = giode.Mix("theme.danger", "theme.surface", 0.14)
	InfoSoft                = giode.Mix("theme.info", "theme.surface", 0.14)
	Selected                = giode.Mix("theme.primary", "theme.surface", 0.10)
	Shadow      giode.Color = "#00000055"
)

// Corner radii.
const (
	RadiusSmall = 6
	Radius      = 10
	RadiusLarge = 16
	RadiusPill  = 999
)

// Motion.
const (
	Quick  = 140 * time.Millisecond
	Smooth = 260 * time.Millisecond
)

// HoverFade animates hover colour changes on boxes.
var HoverFade = giode.StyleTransition{Duration: Quick, Easing: giode.EaseOut}

// Text styles.
var (
	Display  = giode.Styles{FontSize: 26, FontWeight: giode.FontWeightBold, Color: Text}
	Heading  = giode.Styles{FontSize: 17, FontWeight: giode.FontWeightSemiBold, Color: Text}
	Body     = giode.Styles{FontSize: 13.5, Color: Text}
	Strong   = giode.Styles{FontSize: 13.5, FontWeight: giode.FontWeightSemiBold, Color: Text}
	Caption  = giode.Styles{FontSize: 12, Color: Muted}
	Overline = giode.Styles{FontSize: 10.5, FontWeight: giode.FontWeightSemiBold, Color: Muted, TextTransform: giode.TextTransformUppercase}
	Code     = giode.Styles{FontSize: 12.5, FontFamily: Mono, Color: Text}
	CodeDim  = giode.Styles{FontSize: 12, FontFamily: Mono, Color: Muted}
)

// Ellipsis makes a text style end in "…" on one line.
func Ellipsis(st giode.Styles) giode.Styles {
	st.TextOverflow = giode.TextOverflowEllipsis
	st.MaxLines = 1
	return st
}

// Tinted returns a text style recoloured.
func Tinted(st giode.Styles, c giode.Color) giode.Styles {
	st.Color = c
	return st
}
