package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"orchard/ui/theme"
)

// SearchBox is a filter field with a magnifier and a clear button.
type SearchBox struct {
	field *input.Input
	clear *IconAction
}

// NewSearchBox creates a search field. Create it once.
func NewSearchBox(placeholder string) *SearchBox {
	s := &SearchBox{}
	s.field = giode.Input.Text(placeholder, giode.Styles{
		Background: "transparent", BorderColor: "transparent", Color: theme.Text, FontSize: 13,
		Padding: giode.SymmetricInset(4, 7),
	})
	s.clear = NewIconAction("x", "Clear", func() { s.field.SetText("") })
	return s
}

// Query is the current text.
func (s *SearchBox) Query() string { return s.field.Text() }

// Focus puts the caret in the field.
func (s *SearchBox) Focus() { s.field.Focus() }

// Layout draws the field.
func (s *SearchBox) Layout(gtx giode.Context) giode.Dimensions {
	border := theme.Hairline
	if s.field.Focused() {
		border = giode.Mix("theme.primary", "theme.border", 0.6)
	}
	kids := []giode.Element{Glyph("search", 15, theme.Muted), giode.Box(giode.Styles{FlexGrow: 1}, s.field)}
	if s.field.Text() != "" {
		kids = append(kids, s.clear)
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Left: 12, Right: 6}, BorderRadius: theme.Radius,
		Background: theme.Sunken, BorderWidth: 1, BorderColor: border, MinHeight: 38,
	}, kids...).Layout(gtx)
}
