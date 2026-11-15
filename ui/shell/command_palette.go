package shell

import (
	"strings"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"

	"orchard/ui/components"
	"orchard/ui/theme"
)

// Command is an entry of the command palette.
type Command struct {
	Icon  string
	Title string
	Hint  string
	Run   func()
}

// CommandPalette is the Mod+K launcher: type to filter repositories and
// actions, arrows to move, Enter to run.
type CommandPalette struct {
	open     bool
	field    *input.Input
	dlg      *dialog.Dialog
	commands func() []Command
	shown    []Command
	cursor   int
}

// maxResults keeps the palette compact.
const maxResults = 9

// NewCommandPalette creates the palette; commands is read each time it
// filters. Create it once.
func NewCommandPalette(commands func() []Command) *CommandPalette {
	c := &CommandPalette{commands: commands}
	c.field = giode.Input.Text("Type a repository or an action…", giode.Styles{
		Background: "transparent", BorderColor: "transparent", FontSize: 15, Color: theme.Text, Padding: giode.SymmetricInset(4, 10),
	}).Submit(func(string) { c.run(c.cursor) }).
		OnKeys(c.onKey, key.NameUpArrow, key.NameDownArrow)
	c.dlg = giode.Dialog(&c.open, c.content, giode.Styles{
		Background: theme.Surface, BorderRadius: 18, BorderWidth: 1,
		BorderColor: giode.Mix("theme.primary", "theme.border", 0.3),
		BoxShadow:   giode.BoxShadow{Y: 30, Blur: 80, Color: "#000000AA"},
	}).Padding(giode.UniformInset(10))
	return c
}

// Open shows the palette with an empty query.
func (c *CommandPalette) Open() {
	c.field.SetText("")
	c.cursor = 0
	c.open = true
	c.field.Focus()
}

func (c *CommandPalette) onKey(e key.Event) {
	if e.State != key.Press || len(c.shown) == 0 {
		return
	}
	switch e.Name {
	case key.NameUpArrow:
		c.cursor = (c.cursor + len(c.shown) - 1) % len(c.shown)
	case key.NameDownArrow:
		c.cursor = (c.cursor + 1) % len(c.shown)
	}
}

func (c *CommandPalette) run(i int) {
	c.shown = c.filter()
	if i < 0 || i >= len(c.shown) {
		return
	}
	cmd := c.shown[i]
	c.open = false
	giode.AfterLayout(cmd.Run)
}

// filter keeps commands whose title holds every word of the query.
func (c *CommandPalette) filter() []Command {
	words := strings.Fields(strings.ToLower(c.field.Text()))
	var out []Command
	for _, cmd := range c.commands() {
		title := strings.ToLower(cmd.Title + " " + cmd.Hint)
		ok := true
		for _, w := range words {
			if !strings.Contains(title, w) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, cmd)
			if len(out) == maxResults {
				break
			}
		}
	}
	return out
}

func (c *CommandPalette) content() giode.Element {
	c.shown = c.filter()
	if c.cursor >= len(c.shown) {
		c.cursor = 0
	}
	rows := []giode.Element{
		giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(10, 2), Wrap: giode.WrapNoWrap},
			components.Glyph("command", 18, theme.Primary), giode.Box(giode.Styles{FlexGrow: 1}, c.field), components.Kbd("Esc")),
		giode.Divider(giode.Styles{Color: theme.Hairline}),
	}
	for i := range c.shown {
		rows = append(rows, c.row(i))
	}
	if len(c.shown) == 0 {
		rows = append(rows, giode.Box(giode.Styles{Padding: giode.UniformInset(18)}, giode.Text("Nothing matches.", theme.Caption)))
	}
	return giode.Stack(giode.Styles{Width: 560, Gap: 4}, rows...)
}

func (c *CommandPalette) row(i int) giode.Element {
	return giode.Keyed("palette-row", i, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			if i >= len(c.shown) {
				return giode.Box(giode.Styles{})
			}
			cur := c.shown[i]
			if s.Hovered && c.cursor != i {
				c.cursor = i
			}
			bg, fg := giode.Color("transparent"), theme.Muted
			if c.cursor == i {
				bg, fg = theme.Selected, theme.Primary
			}
			return giode.Row(giode.Styles{
				Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(12, 9), BorderRadius: 10, Background: bg,
			},
				components.Glyph(cur.Icon, 16, fg),
				giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(cur.Title, theme.Ellipsis(theme.Body))),
				giode.Text(cur.Hint, theme.Ellipsis(theme.Caption)),
			)
		}).OnClick(func(giode.ClickEvent) { c.run(i) })
	})
}

// Layout draws the palette while open.
func (c *CommandPalette) Layout(gtx giode.Context) giode.Dimensions { return c.dlg.Layout(gtx) }
