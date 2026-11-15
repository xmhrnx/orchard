package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/theme"
)

// PurgePathsCard removes files and folders from every commit: for a
// committed .env, a build folder, or a huge asset.
type PurgePathsCard struct {
	ctx     *rewriteContext
	paths   *input.Input
	largest []gitx.BlobInfo
	sel     *components.Selection
	finding bool
	find    *giode.ButtonElement
	run     *giode.ButtonElement
}

// NewPurgePathsCard creates the card. Create it once.
func NewPurgePathsCard(c *rewriteContext) *PurgePathsCard {
	p := &PurgePathsCard{ctx: c, sel: components.NewSelection()}
	p.paths = giode.Input.Text("secrets.env\nbuild/\nassets/video.mov", components.FieldStyle().With(giode.Styles{FontFamily: theme.Mono, FontSize: 12.5})).Lines(3, 8)
	p.find = components.Button("Find largest files", "search", components.Secondary, p.findLargest)
	p.run = components.Button("Remove from history", "scissors", components.Destructive, p.confirm)
	return p
}

// Reset clears the card for another repository.
func (p *PurgePathsCard) Reset() {
	p.paths.SetText("")
	p.largest = nil
	p.sel.Clear()
}

func (p *PurgePathsCard) findLargest() {
	p.finding = true
	path := p.ctx.repo.Path
	core.Query(p.ctx.env.WS, func(ctx context.Context) ([]gitx.BlobInfo, error) {
		return gitx.LargestFiles(ctx, path, 12)
	}, func(blobs []gitx.BlobInfo, err error) {
		p.finding = false
		p.ctx.env.Fail("Scanning history", err)
		p.largest = blobs
	})
}

func (p *PurgePathsCard) targets() []string {
	seen := map[string]bool{}
	var out []string
	for _, l := range strings.Split(p.paths.Text(), "\n") {
		if l = strings.TrimSpace(l); l != "" && !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	for _, k := range p.sel.Keys() {
		if !seen[k] {
			out = append(out, k)
		}
	}
	return out
}

func (p *PurgePathsCard) confirm() {
	targets := p.targets()
	if len(targets) == 0 {
		p.ctx.env.Toasts.Show(giode.ToastWarning, "List at least one path to remove")
		return
	}
	body := fmt.Sprintf("%s will be erased from every commit on every branch and tag.", strings.Join(targets, ", ")) + rewriteWarning
	p.ctx.guarded("Remove from history?", body, "Rewrite history", "Rewriting history", "History rewritten", func(ctx context.Context, path string) error {
		return gitx.RemovePathsFromHistory(ctx, path, targets)
	})
}

func (p *PurgePathsCard) largestList() giode.Element {
	if p.finding {
		return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary}), giode.Text("Weighing every file ever committed…", theme.Caption))
	}
	if len(p.largest) == 0 {
		return giode.Box(giode.Styles{})
	}
	rows := []giode.Element{giode.Text("Largest files in history: check any to remove", theme.Overline)}
	for _, b := range p.largest {
		rows = append(rows, components.SelectRow("blob", b.Path, p.sel, func() giode.Element {
			return giode.Row(giode.Styles{Width: -1, Gap: 10, Wrap: giode.WrapNoWrap},
				giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(b.Path, theme.Ellipsis(theme.Code))),
				giode.Text(humanSize(b.Size), theme.CodeDim))
		}))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2}, rows...)
}

// Layout draws the card.
func (p *PurgePathsCard) Layout(gtx giode.Context) giode.Dimensions {
	return RewriteCard("scissors", "Remove files from history",
		"Erase files or folders from every commit, as if they were never added. One path per line, relative to the repository root.", false,
		p.paths,
		p.largestList(),
		giode.Row(giode.Styles{Gap: 8, Justify: giode.JustifyEnd, Width: -1}, p.find, p.run),
	).Layout(gtx)
}

func humanSize(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}
