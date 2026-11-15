package screens

import (
	"context"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dropdown"

	"orchard/internal/gitx"
	"orchard/ui/components"
)

var blobSizes = []string{"1M", "5M", "10M", "25M", "50M", "100M"}

// StripBlobsCard drops every file over a size from history, the quick
// way to slim a repository bloated by binaries.
type StripBlobsCard struct {
	ctx  *rewriteContext
	size *dropdown.Select
	run  *giode.ButtonElement
}

// NewStripBlobsCard creates the card. Create it once.
func NewStripBlobsCard(c *rewriteContext) *StripBlobsCard {
	s := &StripBlobsCard{ctx: c}
	s.size = giode.Input.Select(components.FieldStyle().With(giode.Styles{Width: 140}), blobSizes...)
	s.size.SetSelected(2)
	s.run = components.Button("Strip large files", "eraser", components.Destructive, s.confirm)
	return s
}

func (s *StripBlobsCard) confirm() {
	size := s.size.Value()
	body := "Every file bigger than " + size + "B will be removed from all of history." + rewriteWarning
	s.ctx.guarded("Strip files over "+size+"B?", body, "Strip", "Stripping", "Large files stripped", func(ctx context.Context, path string) error {
		return gitx.StripLargeBlobs(ctx, path, size)
	})
}

// Layout draws the card.
func (s *StripBlobsCard) Layout(gtx giode.Context) giode.Dimensions {
	s.run.Disabled(!gitx.HasFilterRepo())
	return RewriteCard("layers", "Strip large files",
		"Remove every file larger than a size from the whole history at once.", true,
		giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			s.size, giode.Box(giode.Styles{FlexGrow: 1}), s.run),
	).Layout(gtx)
}
