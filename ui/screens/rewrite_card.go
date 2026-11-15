package screens

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/components"
	"orchard/ui/theme"
)

// RewriteCard frames one history tool: a glyph, a title, what it does,
// and its controls.
func RewriteCard(icon, title, about string, needsFilterRepo bool, body ...giode.Element) giode.Element {
	head := []giode.Element{
		giode.Box(giode.Styles{Padding: giode.UniformInset(9), BorderRadius: 11, Background: theme.DangerSoft}, components.Glyph(icon, 18, theme.Danger)),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
			giode.Text(title, theme.Heading),
			giode.Text(about, giode.Styles{FontSize: 12.5, Color: theme.Muted, LineHeight: 18}),
		),
	}
	if needsFilterRepo {
		head = append(head, components.Pill("sparkle", "git-filter-repo", theme.Info, theme.InfoSoft))
	}
	kids := append([]giode.Element{giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignStart, Wrap: giode.WrapNoWrap}, head...)}, body...)
	return components.Frame(giode.Styles{Width: -1, Gap: 14, Padding: giode.UniformInset(18)}, kids...)
}
