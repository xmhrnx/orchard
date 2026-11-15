package screens

import (
	"context"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/numberinput"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/segmented"
	"github.com/xmhrnx/giode/components/toggle"

	"orchard/internal/gitx"
	"orchard/internal/store"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

var pullStrategies = []gitx.PullStrategy{gitx.PullFastForward, gitx.PullRebase, gitx.PullMerge}

// SettingsScreen holds the preferences. Every change saves at once.
type SettingsScreen struct {
	env         *env.Env
	themeMode   *segmented.Control
	strategy    *segmented.Control
	autostash   *toggle.Switch
	preferSSH   *toggle.Switch
	concurrency *numberinput.NumberInput
	autoFetch   *numberinput.NumberInput
	scanDepth   *numberinput.NumberInput
	cloneDir    *FolderField
	saveDir     *giode.ButtonElement
	gitVersion  string
	scroll      *scroll.Scroll
}

// NewSettingsScreen creates the page; setDark applies a theme choice.
// Create it once.
func NewSettingsScreen(e *env.Env, setDark func(bool)) *SettingsScreen {
	s := &SettingsScreen{env: e}
	set := func(change func(*store.Settings)) { e.Fail("Saving settings", e.WS.UpdateSettings(change)) }
	cur := e.WS.Settings()

	s.themeMode = giode.Input.Segmented("Dark", "Light").OnChange(func(i int) { setDark(i == 0) })
	s.themeMode.Icons(components.Glyph("moon", 14, ""), components.Glyph("sun", 14, ""))
	s.strategy = giode.Input.Segmented("Fast-forward", "Rebase", "Merge").OnChange(func(i int) {
		set(func(st *store.Settings) { st.PullStrategy = pullStrategies[i] })
	})
	for i, p := range pullStrategies {
		if p == cur.PullStrategy {
			s.strategy.SetSelected(i)
		}
	}
	s.autostash = giode.Input.Switch("").OnChange(func(on bool) { set(func(st *store.Settings) { st.AutoStash = on }) })
	s.autostash.SetChecked(cur.AutoStash)
	s.preferSSH = giode.Input.Switch("").OnChange(func(on bool) { set(func(st *store.Settings) { st.PreferSSH = on }) })
	s.preferSSH.SetChecked(cur.PreferSSH)
	s.concurrency = giode.Input.Number(float64(cur.Concurrency)).Range(1, 16).Step(1).Decimals(0).
		OnChange(func(v float64) { set(func(st *store.Settings) { st.Concurrency = int(v) }) })
	s.autoFetch = giode.Input.Number(float64(cur.AutoFetchMinutes)).Range(0, 240).Step(5).Decimals(0).
		OnChange(func(v float64) { set(func(st *store.Settings) { st.AutoFetchMinutes = int(v) }) })
	s.scanDepth = giode.Input.Number(float64(cur.ScanDepth)).Range(1, 10).Step(1).Decimals(0).
		OnChange(func(v float64) { set(func(st *store.Settings) { st.ScanDepth = int(v) }) })
	s.cloneDir = NewFolderField(e, "Default clone folder")
	s.cloneDir.SetPath(cur.CloneDir)
	s.saveDir = components.Button("Save", "check", components.Secondary, func() {
		dir := s.cloneDir.Path()
		set(func(st *store.Settings) { st.CloneDir = dir })
		e.Toasts.Show(giode.ToastSuccess, "Default clone folder saved")
	})
	s.scroll = giode.Scroll(s.sections, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	go func() {
		v, err := gitx.Version(context.Background())
		e.App.Post(func() {
			s.gitVersion = v
			if err != nil {
				s.gitVersion = "not found: install git and restart Orchard"
			}
		})
	}()
	return s
}

func (s *SettingsScreen) about() giode.Element {
	filterRepo := "not installed (optional, enables every rewrite tool)"
	if gitx.HasFilterRepo() {
		filterRepo = "installed"
	}
	version := giode.AppVersion
	if version == "" {
		version = "development build"
	}
	line := func(k, v string) giode.Element {
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 160}, giode.Text(k, theme.Caption)),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(v, theme.Ellipsis(theme.Code))))
	}
	return components.Section("About", nil,
		line("Orchard", version),
		line("git", s.gitVersion),
		line("git-filter-repo", filterRepo),
		line("Settings file", s.env.WS.StatePath()),
	)
}

func (s *SettingsScreen) sections() giode.Element {
	if theme.IsDark() {
		s.themeMode.SetSelected(0)
	} else {
		s.themeMode.SetSelected(1)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 16, MaxWidth: 880, Padding: giode.Inset{Right: 10, Bottom: 24}},
		components.Section("Appearance", nil,
			SettingRow("Theme", "Night orchard or morning orchard.", s.themeMode),
		),
		components.Section("Updating", nil,
			SettingRow("Pull strategy", "How Update integrates upstream commits. Fast-forward never creates merge commits and stops if branches diverged.", s.strategy),
			SettingRow("Stash local changes", "Set aside uncommitted work while pulling and bring it back after, so dirty repositories still update.", s.autostash),
			SettingRow("Parallel operations", "How many repositories are fetched or pulled at the same time.", s.concurrency),
			SettingRow("Background fetch (minutes)", "Keep incoming-commit counts fresh by fetching every repository on this interval. 0 turns it off.", s.autoFetch),
		),
		components.Section("Cloning", nil,
			components.Field("Default clone folder", "New clones go here unless you choose another folder.",
				giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, giode.Box(giode.Styles{FlexGrow: 1}, s.cloneDir), s.saveDir)),
			SettingRow("Prefer SSH", "Clone account repositories over SSH with your keys instead of HTTPS with the account token.", s.preferSSH),
			SettingRow("Scan depth", "How many folders deep Scan looks for repositories.", s.scanDepth),
		),
		s.about(),
	)
}

// Layout draws the page.
func (s *SettingsScreen) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 22, Padding: giode.Inset{Top: 26, Left: 32, Right: 32}},
		components.PageHeader("Settings", "Make Orchard work the way you do."),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, s.scroll),
	).Layout(gtx)
}
