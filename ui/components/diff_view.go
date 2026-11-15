package components

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/gitx"
	"orchard/ui/theme"
)

// DiffView shows a unified diff in a monospaced, colour-coded list that
// only builds the lines on screen, so huge patches stay smooth.
type DiffView struct {
	lines []gitx.DiffLine
	list  *giode.LazyListElement
}

// NewDiffView creates an empty diff view. Create it once.
func NewDiffView() *DiffView {
	d := &DiffView{}
	d.list = giode.LazyList(
		func() int { return len(d.lines) },
		func(i int) any { return i },
		func(i int) giode.Element { return diffLine(d.lines[i]) },
		giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	).Version(func(i int) any { return d.lines[i] })
	return d
}

// SetLines replaces the diff shown.
func (d *DiffView) SetLines(lines []gitx.DiffLine) {
	d.lines = lines
	d.list.ScrollToKey(0, giode.LazyAlignStart)
}

// Empty reports whether there is nothing to show.
func (d *DiffView) Empty() bool { return len(d.lines) == 0 }

// Layout draws the diff.
func (d *DiffView) Layout(gtx giode.Context) giode.Dimensions { return d.list.Layout(gtx) }

func diffLine(l gitx.DiffLine) giode.Element {
	fg, bg := theme.Text, giode.Color("transparent")
	switch l.Kind {
	case gitx.LineAdded:
		fg, bg = giode.Mix("theme.success", "theme.text", 0.75), giode.Mix("theme.success", "theme.surface", 0.12)
	case gitx.LineRemoved:
		fg, bg = giode.Mix("theme.danger", "theme.text", 0.75), giode.Mix("theme.danger", "theme.surface", 0.12)
	case gitx.LineHunk:
		fg, bg = theme.Info, theme.InfoSoft
	case gitx.LineMeta:
		fg = theme.Muted
	}
	text := l.Text
	if text == "" {
		text = " "
	}
	return giode.Box(giode.Styles{Width: -1, Background: bg, Padding: giode.SymmetricInset(14, 1)},
		giode.Text(text, giode.Styles{FontFamily: theme.Mono, FontSize: 12, Color: fg, LineHeight: 18}))
}
