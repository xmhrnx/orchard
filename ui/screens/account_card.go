package screens

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/theme"
)

// AccountCard is a connected account with its service and a way to
// disconnect it.
type AccountCard struct {
	acc    forge.Account
	remove *components.IconAction
}

// AccountCardFor returns the card of an account, keeping its state.
func AccountCardFor(acc forge.Account, onRemove func(forge.Account)) *AccountCard {
	c := giode.Keyed("account-card", acc.ID, func() *AccountCard {
		card := &AccountCard{acc: acc}
		card.remove = components.NewIconAction("trash", "Disconnect", func() { onRemove(card.acc) })
		return card
	})
	c.acc = acc
	return c
}

// Layout draws the card.
func (c *AccountCard) Layout(gtx giode.Context) giode.Dimensions {
	tint := giode.Color(kindTint(c.acc.Kind))
	return components.Frame(giode.Styles{FlexDirection: giode.FlexDirectionRow, Width: 340, Gap: 14, Align: giode.AlignCenter, Padding: giode.UniformInset(16), Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{Padding: giode.UniformInset(11), BorderRadius: 12, Background: giode.Mix(string(tint), "theme.surface", 0.15)},
			components.Glyph("server", 20, tint)),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 4},
			giode.Text(c.acc.Username, theme.Ellipsis(theme.Strong)),
			giode.Text(c.acc.Host, theme.Ellipsis(theme.CodeDim)),
			components.Pill("", c.acc.Kind.Title(), tint, giode.Mix(string(tint), "theme.surface", 0.12)),
		),
		c.remove,
	).Layout(gtx)
}
