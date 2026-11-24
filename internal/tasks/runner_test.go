package tasks

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"orchard/internal/project"
)

func waitFor(t *testing.T, task *Task) Snapshot {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if s := task.Snapshot(); s.State != Running {
			return s
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("task never finished")
	return Snapshot{}
}

func sh(script string) project.Command { return project.Command{Args: []string{"sh", "-c", script}} }

func TestTaskRunsStepsAndCapturesOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	r := NewRunner(nil)
	task := r.Start("repo", t.TempDir(), "build", "Build", []project.Command{sh("echo one; echo two >&2"), sh("echo \x1b[32mthree\x1b[0m")})
	s := waitFor(t, task)
	if s.State != Succeeded {
		t.Fatalf("state = %v, err %v", s.State, s.Err)
	}
	var out []string
	for _, l := range s.Lines {
		if !l.Info {
			out = append(out, l.Text)
		}
	}
	if strings.Join(out, ",") != "one,two,three" && strings.Join(out, ",") != "two,one,three" {
		t.Errorf("output = %v", out)
	}
}

func TestFailingStepStopsTheTask(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	dir := t.TempDir()
	r := NewRunner(nil)
	task := r.Start("repo", dir, "build", "Build", []project.Command{sh("exit 3"), sh("touch ran")})
	if s := waitFor(t, task); s.State != Failed {
		t.Fatalf("state = %v", s.State)
	}
	if _, err := os.Stat(filepath.Join(dir, "ran")); err == nil {
		t.Error("the step after a failure ran")
	}
}

func TestStopKillsChildProcesses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	dir := t.TempDir()
	r := NewRunner(nil)
	// The shell starts a grandchild that would write a file later.
	task := r.Start("repo", dir, "run", "Run", []project.Command{sh("(sleep 1; touch late) & sleep 30")})
	time.Sleep(200 * time.Millisecond)
	task.Stop()
	if s := waitFor(t, task); s.State != Stopped {
		t.Fatalf("state = %v", s.State)
	}
	time.Sleep(1500 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(dir, "late")); err == nil {
		t.Error("a child process outlived Stop")
	}
}

func TestStartingTheSameKindAgainReplacesIt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	r := NewRunner(nil)
	first := r.Start("repo", t.TempDir(), "run", "Run", []project.Command{sh("sleep 30")})
	time.Sleep(100 * time.Millisecond)
	second := r.Start("repo", t.TempDir(), "run", "Run", []project.Command{sh("sleep 30")})
	if s := waitFor(t, first); s.State != Stopped {
		t.Fatalf("first run = %v, want stopped", s.State)
	}
	if r.Running("repo", "run") != second {
		t.Error("the new run isn't the running one")
	}
	second.Stop()
	waitFor(t, second)
}

func TestMissingProgramFails(t *testing.T) {
	r := NewRunner(nil)
	task := r.Start("repo", t.TempDir(), "build", "Build", []project.Command{{Args: []string{"no-such-tool-orchard"}}})
	if s := waitFor(t, task); s.State != Failed || s.Err == nil {
		t.Fatalf("snapshot = %+v", s)
	}
}
