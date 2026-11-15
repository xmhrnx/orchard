package components

import "github.com/xmhrnx/giode"

// MenuItem is an entry of a RowMenu.
type MenuItem struct {
	Icon  string
	Label string
	Run   func()
}

// rowMenu is the keyed state behind RowMenu.
type rowMenu struct {
	items   []MenuItem
	menu    *giode.MenuElement
	trigger giode.Element
}

// RowMenu is a "⋯" button opening a menu of actions for one list row.
// items may change between frames; the shape (count and labels) is fixed
// when the row first appears.
func RowMenu(scope, key string, items ...MenuItem) giode.Element {
	m := giode.Keyed(scope+"/menu", key, func() *rowMenu {
		r := &rowMenu{menu: giode.Menu()}
		for i, it := range items {
			if it.Label == "" {
				r.menu.Separator()
				continue
			}
			r.menu.IconItem(GlyphSource(it.Icon), it.Label, func() { r.items[i].Run() })
		}
		r.trigger = r.menu.Trigger(NewIconAction("dots", "Actions", func() {}))
		return r
	})
	m.items = items
	return giode.Box(giode.Styles{}, m.trigger, m.menu)
}

// MenuSeparator is a divider between groups of RowMenu items.
var MenuSeparator = MenuItem{}
