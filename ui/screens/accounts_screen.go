package screens

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// AccountsScreen lists the connected hosting accounts.
type AccountsScreen struct {
	env     *env.Env
	dialog  *AccountDialog
	confirm *components.ConfirmDialog
	add     *giode.ButtonElement
	first   *giode.ButtonElement
}

// NewAccountsScreen creates the page. Create it once.
func NewAccountsScreen(e *env.Env) *AccountsScreen {
	s := &AccountsScreen{env: e, dialog: NewAccountDialog(e), confirm: components.NewConfirmDialog()}
	s.add = components.Button("Connect account", "plus", components.Primary, s.dialog.Open)
	s.first = components.Button("Connect your first account", "plus", components.Primary, s.dialog.Open)
	return s
}

func (s *AccountsScreen) remove(acc forge.Account) {
	s.confirm.Ask("Disconnect "+acc.Username+"?", "The token is deleted from this computer. Repositories you cloned stay where they are.", "Disconnect", true, func() {
		giode.AfterLayout(func() { s.env.Fail("Disconnecting", s.env.WS.RemoveAccount(acc.ID)) })
	})
}

func (s *AccountsScreen) note() giode.Element {
	return giode.Row(giode.Styles{
		Width: -1, Gap: 10, Align: giode.AlignCenter, Padding: giode.UniformInset(14), BorderRadius: 12,
		Background: theme.InfoSoft, Wrap: giode.WrapNoWrap,
	},
		components.Glyph("lock", 16, theme.Info),
		giode.Text("Tokens live in your system's credential store (Keychain, Credential Manager or Secret Service), never in Orchard's settings file. Plain git over SSH or HTTPS works without any account; accounts add repository browsing and pull requests.",
			giode.Styles{FontSize: 12.5, Color: theme.Text, LineHeight: 18}),
	)
}

// Layout draws the page.
func (s *AccountsScreen) Layout(gtx giode.Context) giode.Dimensions {
	var body giode.Element
	accounts := s.env.WS.Accounts()
	if len(accounts) == 0 {
		body = components.EmptyState("user", "No accounts connected",
			"Connect GitHub, GitLab, Gitea / Forgejo or Bitbucket, on the public site or your own server.", s.first)
	} else {
		cards := make([]giode.Element, len(accounts))
		for i, a := range accounts {
			cards[i] = AccountCardFor(a, s.remove)
		}
		body = giode.Row(giode.Styles{Width: -1, Gap: 14, Wrap: giode.WrapWrap}, cards...)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 22, Padding: giode.Inset{Top: 26, Left: 32, Right: 32}},
		components.PageHeader("Accounts", "Hosting services you're signed in to.", s.add),
		s.note(),
		body,
		s.dialog, s.confirm,
	).Layout(gtx)
}
