package project

import (
	"os"
	"path/filepath"
	"testing"
)

const doctorOutput = `
Go (every target)
  ✓ go: go1.26.6

Android (giode build android, giode run android)
  ✗ Android SDK: ANDROID_HOME isn't set and no SDK is in the default location
      Install Android Studio (https://developer.android.com/studio), or the command-line tools,
      then set ANDROID_HOME to the SDK directory.
  ✓ JDK (javac, keytool): javac 21.0.12.1

Windows packaging (giode build windows-installer, -release signing)
  ✗ NSIS (makensis)
      winget install NSIS.NSIS, brew install makensis, or apt install nsis

Linux (giode build linux, giode run)
  ✓ Gio's system libraries

3 problem(s) found. Targets that don't need the missing tools still build.
`

func TestParseDoctorAndAvailability(t *testing.T) {
	d := ParseDoctor(doctorOutput)
	if len(d.Sections) != 4 {
		t.Fatalf("sections = %+v", d.Sections)
	}
	want := map[string]bool{
		"linux": true, "windows": true, "web": true, "deb": true,
		"android": false, "aab": false, "windows-installer": false, "macos": false,
	}
	for _, tg := range Targets {
		w, ok := want[tg.ID]
		if !ok {
			continue
		}
		if got := d.availableOn(tg, "linux"); got.OK != w {
			t.Errorf("%s available = %v (%s), want %v", tg.ID, got.OK, got.Reason, w)
		}
	}
	if r := d.availableOn(Targets[10], "linux").Reason; r != "Android SDK: ANDROID_HOME isn't set and no SDK is in the default location — Install Android Studio (https://developer.android.com/studio), or the command-line tools," {
		t.Errorf("android reason = %q", r)
	}
	if r := d.availableOn(Targets[6], "linux").Reason; r != "builds on a Mac only" {
		t.Errorf("macos reason = %q", r)
	}
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetect(t *testing.T) {
	giode := t.TempDir()
	write(t, giode, "giode.toml", "[app]\n")
	write(t, giode, "go.mod", "module x\n")
	if p, ok, _ := Detect(giode); !ok || p.Kind != Giode {
		t.Errorf("giode project detected as %+v", p)
	}

	node := t.TempDir()
	write(t, node, "package.json", `{"scripts":{"dev":"vite","build":"vite build","lint":"eslint ."}}`)
	write(t, node, "pnpm-lock.yaml", "")
	p, ok, err := Detect(node)
	if err != nil || !ok || p.Kind != Node || p.Manager != "pnpm" || !p.NeedsInstall {
		t.Fatalf("node project = %+v, %v, %v", p, ok, err)
	}
	if !p.HasScript("dev") || p.HasScript("start") || p.HasScript("lint") {
		t.Errorf("scripts = %v", p.Scripts)
	}
	steps := p.NodeScript("dev")
	if len(steps) != 2 || steps[0].Args[1] != "install" || steps[1].Args[2] != "dev" {
		t.Errorf("steps = %+v", steps)
	}

	plain := t.TempDir()
	if _, ok, _ := Detect(plain); ok {
		t.Error("an empty folder was detected as a project")
	}
}
