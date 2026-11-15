package screens

import (
	"context"
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/numberinput"

	"orchard/internal/gitx"
	"orchard/ui/components"
)

// SquashCard tidies the tip of the current branch: squash the last few
// commits, reword or undo the latest one.
type SquashCard struct {
	ctx     *rewriteContext
	count   *numberinput.NumberInput
	message *input.Input
	squash  *giode.ButtonElement
	reword  *giode.ButtonElement
	undo    *giode.ButtonElement
	prompt  *components.PromptDialog
}

// NewSquashCard creates the card. Create it once.
func NewSquashCard(c *rewriteContext) *SquashCard {
	s := &SquashCard{ctx: c, prompt: components.NewPromptDialog()}
	s.count = giode.Input.Number(2).Range(2, 200).Step(1).Decimals(0)
	s.message = giode.Input.Text("Message for the combined commit", components.FieldStyle())
	s.squash = components.Button("Squash", "merge", components.Destructive, s.confirmSquash)
	s.reword = components.Button("Reword last commit", "settings", components.Secondary, s.rewordHead)
	s.undo = components.Button("Undo last commit", "undo", components.Secondary, s.confirmUndo)
	return s
}

func (s *SquashCard) confirmSquash() {
	n := int(s.count.Value())
	msg := s.message.Text()
	if msg == "" {
		s.ctx.env.Toasts.Show(giode.ToastWarning, "Write a message for the combined commit")
		return
	}
	body := fmt.Sprintf("The last %d commits of the current branch become one. If they were pushed, you'll need to force-push.", n)
	s.ctx.guarded(fmt.Sprintf("Squash %d commits?", n), body, "Squash", "Squashing", "Commits squashed", func(ctx context.Context, path string) error {
		return gitx.SquashLast(ctx, path, n, msg)
	})
}

func (s *SquashCard) rewordHead() {
	path := s.ctx.repo.Path
	go func() {
		body, err := gitx.CommitBody(context.Background(), path, "HEAD")
		s.ctx.env.App.Post(func() {
			if err != nil {
				s.ctx.env.Fail("Reading last commit", err)
				return
			}
			s.prompt.Ask("Reword last commit", "Message", "", body, "Save", func(msg string) {
				s.ctx.env.WS.Do(s.ctx.repo.ID, "Rewording", func(ctx context.Context, path string) error {
					return gitx.RewordHead(ctx, path, msg)
				}, func(err error) { s.ctx.env.Done("Commit reworded", err) })
			})
		})
	}()
}

func (s *SquashCard) confirmUndo() {
	s.ctx.guarded("Undo last commit?", "The latest commit is removed from the branch; its changes stay staged so nothing is lost.",
		"Undo commit", "Undoing", "Last commit undone", gitx.UndoLastCommit)
}

// Layout draws the card.
func (s *SquashCard) Layout(gtx giode.Context) giode.Dimensions {
	return RewriteCard("merge", "Tidy recent commits",
		"Combine the last commits into one, change the latest message, or step back one commit keeping its changes.", false,
		giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignEnd, Wrap: giode.WrapNoWrap},
			components.Field("Commits", "", s.count),
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("Message", "", s.message)),
			s.squash),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, s.undo, s.reword),
		s.prompt,
	).Layout(gtx)
}
