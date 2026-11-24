package screens

import (
	"regexp"
	"time"

	"github.com/xmhrnx/giode"

	"orchard/internal/tasks"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// consoleHeight is the console's height while open, in dp.
const consoleHeight = 280

// ConsolePanel shows the output of a repository's tasks under its
// page: a chip per task, the chosen task's output following its end,
// and buttons to stop, run again or dismiss it.
type ConsolePanel struct {
	env      *env.Env
	repoID   func() string
	selected map[string]int
	newest   map[string]int
	hidden   map[string]bool
	snap     tasks.Snapshot
	list     *giode.LazyListElement
	stop     *giode.ButtonElement
	again    *giode.ButtonElement
	dismiss  *components.IconAction
	hide     *components.IconAction
}

// NewConsolePanel creates the console. Create it once.
func NewConsolePanel(e *env.Env, repoID func() string) *ConsolePanel {
	c := &ConsolePanel{env: e, repoID: repoID, selected: map[string]int{}, newest: map[string]int{}, hidden: map[string]bool{}}
	c.list = giode.LazyList(
		func() int { return len(c.snap.Lines) },
		func(i int) any { return c.snap.Dropped + i },
		func(i int) giode.Element { return consoleLine(c.snap.Lines[i]) },
		giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	).Anchor(giode.LazyAnchorBottom)
	c.stop = components.Button("Stop", "stop", components.Destructive, func() {
		if t := c.task(); t != nil {
			t.Stop()
		}
	})
	c.again = components.Button("Run again", "fetch", components.Secondary, c.rerun)
	c.dismiss = components.NewIconAction("trash", "Remove this output", func() {
		if t := c.task(); t != nil {
			e.Tasks.Remove(t)
		}
	})
	c.hide = components.NewIconAction("chevron-down", "Hide the console", func() { c.hidden[c.repoID()] = true })
	return c
}

// Show reopens the console, when the user asks for it.
func (c *ConsolePanel) Show() { c.hidden[c.repoID()] = false }

// Visible reports whether the console has anything to show and isn't
// hidden.
func (c *ConsolePanel) Visible() bool {
	return len(c.env.Tasks.Tasks(c.repoID())) > 0 && !c.hidden[c.repoID()]
}

// task is the task shown: the chosen one, else the newest.
func (c *ConsolePanel) task() *tasks.Task {
	list := c.env.Tasks.Tasks(c.repoID())
	if len(list) == 0 {
		return nil
	}
	want := c.selected[c.repoID()]
	for _, t := range list {
		if t.ID == want {
			return t
		}
	}
	return list[len(list)-1]
}

// follow switches to a task as soon as it starts, and reopens the
// console for it.
func (c *ConsolePanel) follow() {
	id := c.repoID()
	list := c.env.Tasks.Tasks(id)
	if len(list) == 0 {
		return
	}
	if newest := list[len(list)-1].ID; newest > c.newest[id] {
		c.newest[id], c.selected[id], c.hidden[id] = newest, newest, false
		c.list.ScrollToBottom()
	}
}

func (c *ConsolePanel) rerun() {
	t := c.task()
	if t == nil {
		return
	}
	c.env.Tasks.Start(t.RepoID, t.Dir, t.Kind, t.Title, t.Steps)
}

func (c *ConsolePanel) chip(t *tasks.Task) giode.Element {
	return giode.Keyed("console-chip", t.ID, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			bg, fg := giode.Color("transparent"), theme.Muted
			if t == c.task() {
				bg, fg = theme.Surface, theme.Text
			} else if s.Hovered {
				bg = theme.Hover
			}
			return giode.Row(giode.Styles{
				Gap: 7, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(10, 5), BorderRadius: 8, Background: bg, Transition: theme.HoverFade,
			}, stateDot(t.State()), giode.Text(t.Title, theme.Ellipsis(giode.Styles{FontSize: 12.5, Color: fg, MaxWidth: 220})))
		}).Label(t.Title).OnClick(func(giode.ClickEvent) {
			c.selected[c.repoID()] = t.ID
			c.list.ScrollToBottom()
		})
	})
}

func stateDot(s tasks.State) giode.Element {
	switch s {
	case tasks.Running:
		return giode.Spinner(giode.Styles{Width: 11, Height: 11, Color: theme.Info})
	case tasks.Succeeded:
		return components.Glyph("check", 12, "theme.success")
	case tasks.Failed:
		return components.Glyph("x", 12, theme.Danger)
	}
	return components.Glyph("stop", 12, theme.Muted)
}

// alarming spots lines that report a problem, whichever stream they
// came on: tools print progress on stderr too, so stderr alone isn't
// a reason to show a line in red.
var alarming = regexp.MustCompile(`(?i)\b(error|errors|failed|fatal|panic)\b`)

func consoleLine(l tasks.Line) giode.Element {
	st := giode.Styles{FontFamily: theme.Mono, FontSize: 12, LineHeight: 18, Color: theme.Text}
	switch {
	case l.Info:
		st.Color = theme.Primary
	case alarming.MatchString(l.Text):
		st.Color = giode.Mix("theme.danger", "theme.text", 0.85)
	case l.Stderr:
		st.Color = theme.Muted
	}
	text := l.Text
	if text == "" {
		text = " "
	}
	return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(14, 0)}, giode.Text(text, st))
}

func (c *ConsolePanel) summary(t *tasks.Task) string {
	end := c.snap.Ended
	if end.IsZero() {
		end = time.Now()
	}
	return c.snap.State.Title() + " · " + components.Duration(end.Sub(t.Started))
}

// Layout draws the console.
func (c *ConsolePanel) Layout(gtx giode.Context) giode.Dimensions {
	c.follow()
	if !c.Visible() {
		return giode.Box(giode.Styles{}).Layout(gtx)
	}
	t := c.task()
	c.snap = t.Snapshot()
	chips := []giode.Element{}
	for _, x := range c.env.Tasks.Tasks(c.repoID()) {
		chips = append(chips, c.chip(x))
	}
	actions := []giode.Element{giode.Text(c.summary(t), theme.Caption)}
	if c.snap.State == tasks.Running {
		actions = append(actions, c.stop)
	} else {
		actions = append(actions, c.again, c.dismiss)
	}
	actions = append(actions, c.hide)
	return components.Frame(giode.Styles{Width: -1, Height: consoleHeight, Padding: giode.UniformInset(0), BorderRadius: 14, Background: theme.Sunken},
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(10, 8)},
			components.Glyph("terminal", 15, theme.Primary),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Row(giode.Styles{Gap: 4, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, chips...)),
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, actions...),
		),
		giode.Divider(giode.Styles{Color: theme.Hairline}),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1, Padding: giode.SymmetricInset(0, 6)}, giode.SelectionArea("console-output", c.list)),
	).Layout(gtx)
}
