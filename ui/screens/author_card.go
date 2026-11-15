package screens

import (
	"context"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"orchard/internal/gitx"
	"orchard/ui/components"
)

// AuthorCard rewrites an identity across history: a work email on a
// personal project, or a typo in a name.
type AuthorCard struct {
	ctx      *rewriteContext
	oldEmail *input.Input
	newName  *input.Input
	newEmail *input.Input
	run      *giode.ButtonElement
}

// NewAuthorCard creates the card. Create it once.
func NewAuthorCard(c *rewriteContext) *AuthorCard {
	a := &AuthorCard{ctx: c}
	a.oldEmail = giode.Input.Text("old@company.com", components.FieldStyle()).Keyboard(giode.KeyboardEmail)
	a.newName = giode.Input.Text("Your Name", components.FieldStyle())
	a.newEmail = giode.Input.Text("you@example.com", components.FieldStyle()).Keyboard(giode.KeyboardEmail)
	a.run = components.Button("Rewrite identity", "user", components.Destructive, a.confirm)
	return a
}

// Reset clears the card for another repository.
func (a *AuthorCard) Reset() {
	for _, f := range []*input.Input{a.oldEmail, a.newName, a.newEmail} {
		f.SetText("")
	}
}

func (a *AuthorCard) confirm() {
	old, name, email := strings.TrimSpace(a.oldEmail.Text()), strings.TrimSpace(a.newName.Text()), strings.TrimSpace(a.newEmail.Text())
	if old == "" || name == "" || email == "" {
		a.ctx.env.Toasts.Show(giode.ToastWarning, "Fill in the old email, the new name and the new email")
		return
	}
	body := "Every commit authored or committed as " + old + " will show " + name + " <" + email + "> instead." + rewriteWarning
	a.ctx.guarded("Rewrite identity?", body, "Rewrite", "Rewriting authors", "Identity rewritten", func(ctx context.Context, path string) error {
		return gitx.RewriteAuthor(ctx, path, old, name, email)
	})
}

// Layout draws the card.
func (a *AuthorCard) Layout(gtx giode.Context) giode.Dimensions {
	return RewriteCard("user", "Fix author identity",
		"Change the name and email on every commit made with a given email address.", false,
		giode.Row(giode.Styles{Width: -1, Gap: 12, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("Old email", "", a.oldEmail)),
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("New name", "", a.newName)),
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("New email", "", a.newEmail)),
		),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd}, a.run),
	).Layout(gtx)
}
