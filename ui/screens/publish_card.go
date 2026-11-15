package screens

import (
	"context"

	"github.com/xmhrnx/giode"

	"orchard/internal/gitx"
	"orchard/ui/components"
)

// PublishCard force-pushes a rewritten history, the last step that
// makes a rewrite reach the server.
type PublishCard struct {
	ctx *rewriteContext
	run *giode.ButtonElement
}

// NewPublishCard creates the card. Create it once.
func NewPublishCard(c *rewriteContext) *PublishCard {
	p := &PublishCard{ctx: c}
	p.run = components.Button("Force-push everything to origin", "push", components.Destructive, p.confirm)
	return p
}

func (p *PublishCard) confirm() {
	p.ctx.guarded("Replace history on origin?",
		"Every branch and tag on origin will be overwritten with your local ones. Commits that exist only on the server are lost.",
		"Force-push", "Force-pushing", "History published", func(ctx context.Context, path string) error {
			return gitx.ForcePushAll(ctx, path, "origin")
		})
}

// Layout draws the card.
func (p *PublishCard) Layout(gtx giode.Context) giode.Dimensions {
	return RewriteCard("push", "Publish the rewrite",
		"A rewrite only changes this computer until you force-push it. Branch protection on the server may need to be lifted first.", false,
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd}, p.run),
	).Layout(gtx)
}
