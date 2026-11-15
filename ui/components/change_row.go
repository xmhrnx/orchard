package components

import (
	"path/filepath"

	"github.com/xmhrnx/giode"

	"orchard/internal/gitx"
	"orchard/ui/theme"
)

// ChangeRow shows a changed file: a coloured status letter, the file
// name and its folder.
func ChangeRow(c gitx.FileChange, selected, hovered bool) giode.Element {
	fg := changeColor(c.Kind)
	bg := giode.Color("transparent")
	switch {
	case selected:
		bg = theme.Selected
	case hovered:
		bg = theme.Hover
	}
	dir := filepath.Dir(c.Path)
	if dir == "." {
		dir = ""
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(10, 6), BorderRadius: 8, Background: bg, Transition: theme.HoverFade,
	},
		giode.Box(giode.Styles{Width: 20, Height: 20, BorderRadius: 5, Background: giode.Mix(string(fg), "theme.surface", 0.16), Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			giode.Text(c.Kind.Letter(), giode.Styles{FontSize: 11, FontWeight: giode.FontWeightBold, FontFamily: theme.Mono, Color: fg})),
		giode.Text(filepath.Base(c.Path), theme.Ellipsis(theme.Body)),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(dir, theme.Ellipsis(theme.Caption))),
	)
}

func changeColor(k gitx.ChangeKind) giode.Color {
	switch k {
	case gitx.Added, gitx.Untracked:
		return "theme.success"
	case gitx.Deleted:
		return theme.Danger
	case gitx.Renamed, gitx.Copied:
		return theme.Info
	case gitx.Conflicted:
		return theme.Warning
	}
	return theme.Accent
}
