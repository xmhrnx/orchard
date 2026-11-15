package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"orchard/ui/theme"
)

// CheckRow is a list row with a checkbox bound to a Selection. Clicking
// anywhere on the row toggles it, as in a mail client.
type CheckRow struct {
	key     string
	sel     *Selection
	box     *checkbox.Checkbox
	press   *giode.PressableElement
	content func() giode.Element
}

// SelectRow returns the row for key, keeping its state across frames.
// content is rebuilt every frame.
func SelectRow(scope, key string, sel *Selection, content func() giode.Element) *CheckRow {
	r := giode.Keyed(scope, key, func() *CheckRow { return newCheckRow(key, sel) })
	r.sel, r.content = sel, content
	return r
}

func newCheckRow(key string, sel *Selection) *CheckRow {
	r := &CheckRow{key: key, sel: sel}
	r.box = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
		OnChange(func(on bool) { r.sel.Set(r.key, on) })
	r.press = giode.Pressable(func(s giode.PressState) giode.Element {
		bg := giode.Color("transparent")
		switch {
		case r.sel.Has(r.key):
			bg = theme.Selected
		case s.Hovered:
			bg = theme.Hover
		}
		return giode.Row(giode.Styles{
			Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(12, 9), BorderRadius: theme.Radius, Background: bg,
			Transition: theme.HoverFade,
		}, r.box, giode.Box(giode.Styles{FlexGrow: 1}, r.content()))
	}).Label(key).OnClick(func(giode.ClickEvent) { r.sel.Toggle(r.key) })
	return r
}

// Layout draws the row.
func (r *CheckRow) Layout(gtx giode.Context) giode.Dimensions {
	r.box.SetChecked(r.sel.Has(r.key))
	return r.press.Layout(gtx)
}
