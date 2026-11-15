package components

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/gitx"
	"orchard/ui/theme"
)

// BranchLabel shows the checked-out branch, or the short commit when the
// head is detached.
func BranchLabel(st gitx.Status) giode.Element {
	name, icon := st.Branch, "branch"
	if st.Detached {
		name, icon = "detached "+short(st.Head), "commit"
	}
	if name == "" {
		name = "—"
	}
	return giode.Row(giode.Styles{
		Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(9, 4), BorderRadius: theme.RadiusSmall,
		Background: theme.Sunken, BorderWidth: 1, BorderColor: theme.Hairline, MaxWidth: 200,
	},
		Glyph(icon, 13, theme.Primary),
		giode.Text(name, theme.Ellipsis(giode.Styles{FontSize: 12, FontFamily: theme.Mono, Color: theme.Text})),
	)
}

func short(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
