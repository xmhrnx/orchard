package screens

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/segmented"

	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// AccountDialog connects a hosting account with an access token, which
// is verified and then kept in the system's credential store.
type AccountDialog struct {
	env      *env.Env
	open     bool
	kind     *segmented.Control
	host     *input.Input
	username *input.Input
	token    *input.Input
	getToken *giode.ButtonElement
	connect  *giode.ButtonElement
	cancel   *giode.ButtonElement
	dlg      *dialog.Dialog
}

// NewAccountDialog creates the dialog. Create it once.
func NewAccountDialog(e *env.Env) *AccountDialog {
	d := &AccountDialog{env: e}
	labels := make([]string, len(forge.Kinds))
	for i, k := range forge.Kinds {
		labels[i] = k.Title()
	}
	d.kind = giode.Input.Segmented(labels...).OnChange(func(i int) { d.host.SetText(forge.DefaultHost(forge.Kinds[i])) })
	d.kind.Styles(giode.Styles{Width: -1})
	d.host = giode.Input.Text("github.com", components.FieldStyle()).Keyboard(giode.KeyboardURL)
	d.username = giode.Input.Text("optional for GitHub, GitLab and Gitea", components.FieldStyle())
	d.token = giode.Input.Text("paste the token", components.FieldStyle()).Password(true).Submit(func(string) { d.submit() })
	d.getToken = components.Button("Create a token", "external", components.Quiet, func() { e.OpenURL(tokenURL(d.selected(), d.host.Text())) })
	d.connect = components.Button("Connect", "check", components.Primary, d.submit)
	d.cancel = components.Button("Cancel", "", components.Secondary, func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.content, components.DialogPanel())
	return d
}

// Open shows an empty form.
func (d *AccountDialog) Open() {
	d.kind.SetSelected(0)
	d.host.SetText(forge.DefaultHost(forge.GitHub))
	d.username.SetText("")
	d.token.SetText("")
	d.connect.Loading(false)
	d.open = true
	d.token.Focus()
}

func (d *AccountDialog) selected() forge.Kind { return forge.Kinds[d.kind.Selected()] }

func (d *AccountDialog) submit() {
	token := strings.TrimSpace(d.token.Text())
	if token == "" {
		d.env.Toasts.Show(giode.ToastWarning, "Paste an access token first")
		return
	}
	d.connect.Loading(true)
	d.env.WS.AddAccount(d.selected(), d.host.Text(), d.username.Text(), token, func(err error) {
		d.connect.Loading(false)
		if err != nil {
			d.env.Fail("Connecting", err)
			return
		}
		d.open = false
		d.env.Toasts.Show(giode.ToastSuccess, "Account connected")
	})
}

func (d *AccountDialog) content() giode.Element {
	return giode.Stack(giode.Styles{Gap: 16, Width: 520},
		giode.Text("Connect an account", theme.Heading),
		d.kind,
		components.Field("Host", "Change it for self-hosted servers, e.g. git.company.com", d.host),
		components.Field("Username", "", d.username),
		components.Field("Access token", tokenHint(d.selected()), d.token),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter},
			d.getToken, giode.Box(giode.Styles{FlexGrow: 1}), d.cancel, d.connect),
	)
}

// Layout draws the dialog.
func (d *AccountDialog) Layout(gtx giode.Context) giode.Dimensions { return d.dlg.Layout(gtx) }
