package components

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"orchard/internal/gitx"
	"orchard/internal/graph"
	"orchard/ui/theme"
)

// CommitRowState is how a history row is shown.
type CommitRowState struct {
	Head     bool
	Selected bool
	Hovered  bool
	// Compact drops the author column and shows fewer ref badges, for
	// when a detail panel shares the width.
	Compact bool
}

// CommitRow is one line of the history: graph, refs, subject, author,
// hash and age. Its background reflects selection and hover.
func CommitRow(c gitx.Commit, row graph.Row, lanes int, st CommitRowState) giode.Element {
	head, selected, hovered := st.Head, st.Selected, st.Hovered
	maxRefBadges := 3
	if st.Compact {
		maxRefBadges = 1
	}
	bg := giode.Color("transparent")
	switch {
	case selected:
		bg = theme.Selected
	case hovered:
		bg = theme.Hover
	}
	subject := []giode.Element{}
	for i, r := range c.Refs {
		if i == maxRefBadges {
			subject = append(subject, Pill("", fmt.Sprintf("+%d", len(c.Refs)-i), theme.Muted, theme.Hover))
			break
		}
		subject = append(subject, RefBadge(r))
	}
	subjectStyle := theme.Body
	if head {
		subjectStyle = theme.Strong
	}
	subject = append(subject, giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(c.Subject, theme.Ellipsis(subjectStyle))))
	cols := []giode.Element{
		GraphCell(row, lanes, head, c.IsMerge()),
		giode.Row(giode.Styles{FlexGrow: 1, Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, subject...),
	}
	if !st.Compact {
		cols = append(cols, giode.Box(giode.Styles{Width: 150, Padding: giode.Inset{Left: 12}}, giode.Text(c.Author, theme.Ellipsis(theme.Caption))))
	}
	cols = append(cols,
		giode.Box(giode.Styles{Width: 84, Padding: giode.Inset{Left: 10}}, giode.Text(c.Short(), theme.CodeDim)),
		giode.Box(giode.Styles{Width: 76}, giode.Text(Ago(c.When), giode.Styles{FontSize: 12, Color: theme.Muted, TextAlign: giode.TextAlignEnd, Width: -1})),
	)
	return giode.Row(giode.Styles{
		Width: -1, Height: GraphRowH, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Right: 14}, Background: bg, BorderRadius: 8, Transition: theme.HoverFade,
	}, cols...)
}
