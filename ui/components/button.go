package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// ButtonKind is a button's visual weight.
type ButtonKind int

// Button kinds, from loudest to quietest.
const (
	Primary ButtonKind = iota
	Secondary
	Quiet
	Destructive
)

// Button makes a themed button with an optional leading glyph. Create
// it once and lay it out every frame.
func Button(label, icon string, kind ButtonKind, onClick func()) *giode.ButtonElement {
	st := giode.Styles{BorderRadius: theme.Radius, FontSize: 13, Padding: giode.SymmetricInset(14, 8)}
	variant := giode.VariantContained
	switch kind {
	case Primary:
		st.BackgroundGradient = giode.Linear(135, "theme.primary", giode.Mix("theme.primary", "theme.secondary", 0.45))
		st.Color = theme.OnPrimary
	case Secondary:
		st.Background, st.Color = theme.Raised, theme.Text
		st.BorderColor, st.BorderWidth = theme.Hairline, 1
	case Quiet:
		st.Background, st.Color = theme.Text, theme.Text
		variant = giode.VariantGhost
	case Destructive:
		st.Background, st.Color = theme.DangerSoft, theme.Danger
		st.BorderColor, st.BorderWidth = giode.Mix("theme.danger", "theme.surface", 0.35), 1
	}
	b := giode.Button(label, st).Variant(variant).OnClick(onClick)
	if icon != "" {
		b.Icon(Glyph(icon, 15, ""))
	}
	return b
}
