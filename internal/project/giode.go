package project

import (
	"bufio"
	"regexp"
	"runtime"
	"strings"
)

// Target is a platform giode build can produce.
type Target struct {
	ID    string
	Title string
	// OS limits the target to building on one system ("darwin", "linux"),
	// or "" for any.
	OS string
	// Needs names the doctor section that must pass, when the target
	// needs more than Go.
	Needs string
}

// Targets is what the build dropdown offers.
var Targets = []Target{
	{ID: "linux", Title: "Linux executable", OS: "linux", Needs: "Linux"},
	{ID: "deb", Title: "Debian package", OS: "linux", Needs: "Linux"},
	{ID: "rpm", Title: "RPM package", OS: "linux", Needs: "Linux"},
	{ID: "appimage", Title: "AppImage", OS: "linux", Needs: "Linux"},
	{ID: "windows", Title: "Windows executable"},
	{ID: "windows-installer", Title: "Windows installer", Needs: "Windows packaging"},
	{ID: "macos", Title: "macOS app", OS: "darwin", Needs: "macOS"},
	{ID: "dmg", Title: "macOS disk image", OS: "darwin", Needs: "macOS"},
	{ID: "ios-sim", Title: "iOS simulator app", OS: "darwin", Needs: "iOS"},
	{ID: "ios", Title: "iOS app", OS: "darwin", Needs: "iOS"},
	{ID: "android", Title: "Android APK", Needs: "Android"},
	{ID: "aab", Title: "Android App Bundle", Needs: "Android"},
	{ID: "web", Title: "Web (WebAssembly)"},
}

// DoctorSection is one block of giode doctor's report.
type DoctorSection struct {
	Title    string
	Problems []string
}

// Doctor is giode doctor's report, by section.
type Doctor struct {
	Sections []DoctorSection
}

// Availability says whether a target can be built here, and why not.
type Availability struct {
	OK     bool
	Reason string
}

var sectionHeader = regexp.MustCompile(`^(\S[^(]*?)\s*(\(.*\))?$`)

// ParseDoctor reads giode doctor's output: unindented section titles,
// then indented "✓" and "✗" checks with hints under failing ones.
func ParseDoctor(out string) Doctor {
	var d Doctor
	hint := false
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
		case !strings.HasPrefix(line, " "):
			if m := sectionHeader.FindStringSubmatch(trimmed); m != nil && !strings.Contains(trimmed, "problem(s) found") {
				d.Sections = append(d.Sections, DoctorSection{Title: m[1]})
			}
		case strings.HasPrefix(trimmed, "✗") && len(d.Sections) > 0:
			s := &d.Sections[len(d.Sections)-1]
			s.Problems = append(s.Problems, strings.TrimSpace(strings.TrimPrefix(trimmed, "✗")))
			hint = true
		case hint && strings.HasPrefix(line, "      ") && len(d.Sections) > 0:
			// The first line under a failed check says how to fix it.
			s := &d.Sections[len(d.Sections)-1]
			s.Problems[len(s.Problems)-1] += " — " + trimmed
			hint = false
			continue
		}
		if !strings.HasPrefix(trimmed, "✗") {
			hint = false
		}
	}
	return d
}

// section finds the section whose title starts with name.
func (d Doctor) section(name string) (DoctorSection, bool) {
	for _, s := range d.Sections {
		if strings.HasPrefix(s.Title, name) {
			return s, true
		}
	}
	return DoctorSection{}, false
}

// Available reports whether t builds on this computer, judging by the
// doctor report.
func (d Doctor) Available(t Target) Availability {
	return d.availableOn(t, runtime.GOOS)
}

func (d Doctor) availableOn(t Target, goos string) Availability {
	if t.OS != "" && t.OS != goos {
		return Availability{Reason: "builds on " + osName(t.OS) + " only"}
	}
	if s, ok := d.section("Go"); ok && len(s.Problems) > 0 {
		return Availability{Reason: s.Problems[0]}
	}
	if t.Needs != "" {
		if s, ok := d.section(t.Needs); ok && len(s.Problems) > 0 {
			return Availability{Reason: s.Problems[0]}
		}
	}
	return Availability{OK: true}
}

// DefaultTargets is what to build before the user picks: this
// computer's own desktop target.
func DefaultTargets() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"macos"}
	case "windows":
		return []string{"windows"}
	}
	return []string{"linux"}
}

func osName(goos string) string {
	switch goos {
	case "darwin":
		return "a Mac"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	}
	return goos
}

// GiodeRun builds and runs the app on this computer.
func GiodeRun() []Command { return []Command{{Args: []string{"giode", "run"}}} }

// GiodeBuild builds the chosen targets in one invocation.
func GiodeBuild(targets []string) []Command {
	return []Command{{Args: append([]string{"giode", "build"}, targets...)}}
}

// GiodeTool runs another giode command (doctor, upgrade, clean, ...).
func GiodeTool(args ...string) []Command {
	return []Command{{Args: append([]string{"giode"}, args...)}}
}
