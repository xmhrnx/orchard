// Package shell is the app chrome around the screens: the custom window
// frame, the sidebar and the command palette.
package shell

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/titlebar"

	"orchard/ui/components"
	"orchard/ui/theme"
)

// WindowBar is the frameless window's title bar: drag area, the current
// page's name in the middle, and quick controls by the window buttons.
type WindowBar struct {
	bar *titlebar.TitleBar
}

// NewWindowBar creates the bar. Create it once.
func NewWindowBar(app *giode.App, openPalette, toggleTheme func()) *WindowBar {
	search := giode.Pressable(func(s giode.PressState) giode.Element {
		border := theme.Hairline
		if s.Hovered {
			border = giode.Mix("theme.primary", "theme.border", 0.4)
		}
		return giode.Row(giode.Styles{
			Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Width: 260,
			Padding: giode.SymmetricInset(10, 4), BorderRadius: 8, Background: theme.Sunken,
			BorderWidth: 1, BorderColor: border, Transition: theme.HoverFade,
		},
			components.Glyph("search", 13, theme.Muted),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Jump to repo or action…", giode.Styles{FontSize: 12, Color: theme.Muted})),
			components.Kbd("Mod+K"),
		)
	}).Label("Command palette").OnClick(func(giode.ClickEvent) { openPalette() })
	themeBtn := components.NewIconAction("sun", "Toggle light / dark", toggleTheme)
	bar := giode.TitleBar(app, "Orchard", giode.Styles{
		Background: theme.Canvas, Color: theme.Muted, Height: 42, FontSize: 12.5,
		BorderWidth: 1, BorderColor: giode.Mix("theme.border", "theme.background", 0.4),
	}).
		Icon(giode.Box(giode.Styles{Padding: giode.Inset{Left: 4}}, components.LogoMark(20))).
		Leading(giode.Box(giode.Styles{Padding: giode.Inset{Left: 8}}, search)).
		Trailing(themeBtn).
		Variant(windowControls).
		TitlePlacement(giode.TitleBarTitleCenter).
		TitleMaxWidth(320)
	return &WindowBar{bar: bar}
}

// SetTitle shows the current page in the bar.
func (w *WindowBar) SetTitle(title string) { w.bar.Title(title) }

// Layout draws the bar.
func (w *WindowBar) Layout(gtx giode.Context) giode.Dimensions { return w.bar.Layout(gtx) }
