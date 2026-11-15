package screens

import (
	"context"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// NewReleaseDialog drafts or publishes a release on the hosting service.
type NewReleaseDialog struct {
	env        *env.Env
	open       bool
	account    forge.Account
	slug       string
	caps       forge.ReleaseCaps
	onCreated  func()
	tag        *input.Input
	target     *input.Input
	name       *input.Input
	notes      *input.Input
	draft      *checkbox.Checkbox
	prerelease *checkbox.Checkbox
	generate   *checkbox.Checkbox
	create     *giode.ButtonElement
	cancel     *giode.ButtonElement
	form       *scroll.Scroll
	dlg        *dialog.Dialog
}

// NewNewReleaseDialog creates the dialog; created runs after a release
// was made. Create it once.
func NewNewReleaseDialog(e *env.Env, created func()) *NewReleaseDialog {
	d := &NewReleaseDialog{env: e, onCreated: created}
	mono := components.FieldStyle().With(giode.Styles{FontFamily: theme.Mono})
	d.tag = giode.Input.Text("v1.2.0", mono)
	d.target = giode.Input.Text("main", mono)
	d.name = giode.Input.Text("same as the tag", components.FieldStyle())
	d.notes = giode.Input.Text("What's new, in Markdown", components.FieldStyle()).Lines(5, 12)
	box := giode.Styles{Background: theme.Primary, BorderColor: theme.Muted, Color: theme.Text}
	d.draft = giode.Input.Checkbox("Save as draft", box)
	d.prerelease = giode.Input.Checkbox("Pre-release", box)
	d.generate = giode.Input.Checkbox("Generate notes from merged pull requests", box)
	d.create = components.Button("Publish release", "release", components.Primary, d.submit)
	d.cancel = components.Button("Cancel", "", components.Secondary, func() { d.open = false })
	d.form = giode.Scroll(d.fields).FitContent(true)
	d.dlg = giode.Dialog(&d.open, d.content, components.DialogPanel())
	return d
}

// Open shows an empty form for a hosted repository; target is the
// branch new tags are made from.
func (d *NewReleaseDialog) Open(acc forge.Account, slug, target string, caps forge.ReleaseCaps) {
	d.account, d.slug, d.caps, d.open = acc, slug, caps, true
	for _, f := range []*input.Input{d.tag, d.name, d.notes} {
		f.SetText("")
	}
	d.target.SetText(target)
	for _, c := range []*checkbox.Checkbox{d.draft, d.prerelease, d.generate} {
		c.SetChecked(false)
	}
	d.create.Loading(false)
	d.tag.Focus()
}

func (d *NewReleaseDialog) submit() {
	rel := forge.NewRelease{
		Tag: strings.TrimSpace(d.tag.Text()), Target: strings.TrimSpace(d.target.Text()),
		Name: strings.TrimSpace(d.name.Text()), Notes: d.notes.Text(),
		Draft: d.draft.Checked(), Prerelease: d.prerelease.Checked(), GenerateNotes: d.generate.Checked(),
	}
	if rel.Tag == "" {
		d.env.Toasts.Show(giode.ToastWarning, "Give the release a tag")
		d.tag.Focus()
		return
	}
	if rel.Name == "" {
		rel.Name = rel.Tag
	}
	acc, slug := d.account, d.slug
	d.create.Loading(true)
	core.Query(d.env.WS, func(ctx context.Context) (struct{}, error) {
		client, err := d.env.WS.Client(acc)
		if err != nil {
			return struct{}{}, err
		}
		rs, ok := forge.ReleasesOf(client)
		if !ok {
			return struct{}{}, forge.ErrNoReleases
		}
		return struct{}{}, rs.CreateRelease(ctx, slug, rel)
	}, func(_ struct{}, err error) {
		d.create.Loading(false)
		if err != nil {
			d.env.Fail("Creating the release", err)
			return
		}
		d.open = false
		what := "Published "
		if rel.Draft {
			what = "Drafted "
		}
		d.env.Toasts.Show(giode.ToastSuccess, what+rel.Name)
		d.onCreated()
	})
}

func (d *NewReleaseDialog) fields() giode.Element {
	var options []giode.Element
	if d.caps.Drafts {
		options = append(options, d.draft)
	}
	if d.caps.Prereleases {
		options = append(options, d.prerelease)
	}
	if d.caps.GenerateNotes {
		options = append(options, d.generate)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 14},
		giode.Row(giode.Styles{Width: -1, Gap: 12, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("Tag", "An existing tag, or a new one", d.tag)),
			giode.Box(giode.Styles{FlexGrow: 1}, components.Field("From", "Branch or commit for a new tag", d.target)),
		),
		components.Field("Title", "", d.name),
		components.Field("Notes", "", d.notes),
		giode.Row(giode.Styles{Width: -1, Gap: 20, Wrap: giode.WrapWrap}, options...),
	)
}

func (d *NewReleaseDialog) content() giode.Element {
	if d.draft.Checked() {
		d.create.Label("Save draft")
	} else {
		d.create.Label("Publish release")
	}
	return giode.Stack(giode.Styles{Gap: 16, Width: 560},
		giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter},
			components.Glyph("release", 20, theme.Primary), giode.Text("New release", theme.Heading)),
		giode.Box(giode.Styles{Width: -1, MaxHeight: 520}, d.form),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, d.create),
	)
}

// Layout draws the dialog.
func (d *NewReleaseDialog) Layout(gtx giode.Context) giode.Dimensions { return d.dlg.Layout(gtx) }
