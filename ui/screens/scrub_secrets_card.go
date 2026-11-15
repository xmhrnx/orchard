package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/theme"
)

// ScrubSecretsCard replaces leaked passwords and keys everywhere in
// history with ***REMOVED***.
type ScrubSecretsCard struct {
	ctx     *rewriteContext
	secrets *input.Input
	run     *giode.ButtonElement
}

// NewScrubSecretsCard creates the card. Create it once.
func NewScrubSecretsCard(c *rewriteContext) *ScrubSecretsCard {
	s := &ScrubSecretsCard{ctx: c}
	s.secrets = giode.Input.Text("ghp_aBcD…\nhunter2", components.FieldStyle().With(giode.Styles{FontFamily: theme.Mono, FontSize: 12.5})).Lines(2, 6).Password(false)
	s.run = components.Button("Scrub secrets", "eraser", components.Destructive, s.confirm)
	return s
}

// Reset clears the card for another repository.
func (s *ScrubSecretsCard) Reset() { s.secrets.SetText("") }

func (s *ScrubSecretsCard) confirm() {
	var secrets []string
	for _, l := range strings.Split(s.secrets.Text(), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			secrets = append(secrets, l)
		}
	}
	if len(secrets) == 0 {
		s.ctx.env.Toasts.Show(giode.ToastWarning, "Enter the text to scrub, one per line")
		return
	}
	body := fmt.Sprintf("%d strings will be replaced with ***REMOVED*** in every file of every commit.", len(secrets)) +
		" Rotate the leaked credentials too: anyone who cloned before still has them." + rewriteWarning
	s.ctx.guarded("Scrub secrets from history?", body, "Scrub", "Scrubbing", "Secrets scrubbed", func(ctx context.Context, path string) error {
		return gitx.ReplaceText(ctx, path, secrets)
	})
}

// Layout draws the card.
func (s *ScrubSecretsCard) Layout(gtx giode.Context) giode.Dimensions {
	s.run.Disabled(!gitx.HasFilterRepo())
	return RewriteCard("key", "Scrub leaked secrets",
		"Replace API keys, tokens or passwords with ***REMOVED*** in every version of every file. One secret per line.", true,
		s.secrets,
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd}, s.run),
	).Layout(gtx)
}
