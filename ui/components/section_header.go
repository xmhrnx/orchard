package components

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// SectionInfo is what a list section header shows.
type SectionInfo struct {
	Icon      string
	Title     string
	Count     int
	Behind    int
	Collapsed bool
}

// SectionHeader titles a section of the repository list: click to fold
// it away, with a quick update for everything in it.
type SectionHeader struct {
	info     SectionInfo
	onToggle func()
	onUpdate func()
	press    *giode.PressableElement
	update   *IconAction
}

// SectionHeaderFor returns the header of section key, keeping its state.
// The callbacks may change between frames.
func SectionHeaderFor(key string, info SectionInfo, onToggle, onUpdate func()) *SectionHeader {
	h := giode.Keyed("section-header", key, func() *SectionHeader {
		h := &SectionHeader{}
		h.update = NewIconAction("pull", "Update this section", func() { h.onUpdate() })
		h.press = giode.Pressable(h.content).Label(key).OnClick(func(giode.ClickEvent) { h.onToggle() })
		return h
	})
	h.info, h.onToggle, h.onUpdate = info, onToggle, onUpdate
	return h
}

func (h *SectionHeader) content(s giode.PressState) giode.Element {
	i := h.info
	chevron, fg := "chevron-down", theme.Text
	if i.Collapsed {
		chevron = "chevron-right"
	}
	if s.Hovered {
		fg = theme.Primary
	}
	kids := []giode.Element{
		Glyph(chevron, 14, theme.Muted),
		Glyph(i.Icon, 15, theme.Primary),
		giode.Text(i.Title, giode.Styles{FontSize: 13.5, FontWeight: giode.FontWeightSemiBold, Color: fg}),
		Pill("", fmt.Sprint(i.Count), theme.Muted, theme.Hover),
	}
	if i.Behind > 0 {
		kids = append(kids, Pill("arrow-down", fmt.Sprintf("%d need updating", i.Behind), theme.Info, theme.InfoSoft))
	}
	kids = append(kids,
		giode.Box(giode.Styles{FlexGrow: 1, Height: 1, Background: theme.Hairline, Margin: giode.SymmetricInset(6, 0)}),
		h.update,
	)
	return giode.Row(giode.Styles{
		Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Top: 10, Bottom: 6, Left: 6, Right: 6},
	}, kids...)
}

// Layout draws the header.
func (h *SectionHeader) Layout(gtx giode.Context) giode.Dimensions { return h.press.Layout(gtx) }
