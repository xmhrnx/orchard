package screens

import (
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// ReleaseActions are what the release panel can ask its view to do.
type ReleaseActions struct {
	Publish func(forge.Release)
	Delete  func(forge.Release)
	Close   func()
}

// ReleasePanel shows one release beside the list: its notes rendered as
// Markdown, its downloads and its actions.
type ReleasePanel struct {
	env     *env.Env
	release forge.Release
	act     ReleaseActions
	body    *scroll.Scroll
	close   *components.IconAction
	open    *components.IconAction
	publish *giode.ButtonElement
	remove  *giode.ButtonElement
}

// NewReleasePanel creates the panel. Create it once.
func NewReleasePanel(e *env.Env, act ReleaseActions) *ReleasePanel {
	p := &ReleasePanel{env: e, act: act}
	p.close = components.NewIconAction("x", "Close", act.Close)
	p.open = components.NewIconAction("external", "Open in browser", func() { e.OpenURL(p.release.WebURL) })
	p.publish = components.Button("Publish", "release", components.Primary, func() { act.Publish(p.release) })
	p.remove = components.Button("Delete", "trash", components.Destructive, func() { act.Delete(p.release) })
	p.body = giode.Scroll(p.content, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	return p
}

// Show puts a release in the panel.
func (p *ReleasePanel) Show(r forge.Release) {
	p.release = r
	p.body.ScrollTo(0)
}

// ID is the release shown.
func (p *ReleasePanel) ID() string { return p.release.ID }

func (p *ReleasePanel) asset(a forge.ReleaseAsset) giode.Element {
	meta := []string{}
	if a.Size > 0 {
		meta = append(meta, humanSize(a.Size))
	}
	if a.Downloads > 0 {
		meta = append(meta, fmt.Sprintf("%d downloads", a.Downloads))
	}
	return giode.Keyed("release-asset", p.release.ID+"/"+a.Name, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			bg := theme.Raised
			if s.Hovered {
				bg = theme.Hover
			}
			return giode.Row(giode.Styles{
				Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(12, 9), BorderRadius: theme.Radius, Background: bg, Transition: theme.HoverFade,
			},
				components.Glyph("download", 15, theme.Primary),
				giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(a.Name, theme.Ellipsis(theme.Body))),
				giode.Text(strings.Join(meta, " · "), theme.Caption),
			)
		}).Label("Download " + a.Name).OnClick(func(giode.ClickEvent) { p.env.OpenURL(a.URL) })
	})
}

func (p *ReleasePanel) content() giode.Element {
	r := p.release
	notes := r.Notes
	if strings.TrimSpace(notes) == "" {
		notes = "_No release notes._"
	}
	kids := []giode.Element{
		giode.Markdown(notes, giode.MarkdownOptions{
			Styles:         giode.Styles{FontSize: 13.5, LineHeight: 20, Color: theme.Text},
			CodeFont:       theme.Mono,
			CodeBackground: theme.Sunken,
			OnLink:         p.env.OpenURL,
			Selectable:     "release-notes",
		}),
	}
	if len(r.Assets) > 0 {
		kids = append(kids, giode.Box(giode.Styles{Padding: giode.Inset{Top: 8}}, giode.Text(fmt.Sprintf("Assets · %d", len(r.Assets)), theme.Overline)))
		for _, a := range r.Assets {
			kids = append(kids, p.asset(a))
		}
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8, Padding: giode.Inset{Right: 8}}, kids...)
}

// Layout draws the panel.
func (p *ReleasePanel) Layout(gtx giode.Context) giode.Dimensions {
	r := p.release
	actions := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1})}
	if r.Draft {
		actions = append(actions, p.publish)
	}
	actions = append(actions, p.remove)
	by := "by " + r.Author
	if r.Author == "" {
		by = ""
	}
	when := r.Published
	if when.IsZero() {
		when = r.Created
	}
	return components.Frame(giode.Styles{Width: 460, Height: -1, Gap: 12, Padding: giode.UniformInset(18)},
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, components.Pill("tag", r.Tag, theme.Accent, theme.AccentSoft)),
			p.open, p.close),
		giode.Text(releaseTitle(r), theme.Ellipsis(theme.Heading)),
		giode.Text(strings.TrimSpace(by+" · "+components.Ago(when)), theme.Caption),
		giode.Divider(giode.Styles{Color: theme.Hairline}),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, p.body),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter}, actions...),
	).Layout(gtx)
}

func releaseTitle(r forge.Release) string {
	if r.Name != "" {
		return r.Name
	}
	return r.Tag
}
