package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// BrandMark is the app's logo with its name and tagline.
func BrandMark() giode.Element {
	return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		LogoMark(34),
		giode.Stack(giode.Styles{Gap: 1},
			giode.Text("Orchard", giode.Styles{FontSize: 17, FontWeight: giode.FontWeightBold, Color: theme.Text}),
			giode.Text("tend every repo", giode.Styles{FontSize: 11, Color: theme.Muted}),
		),
	)
}
