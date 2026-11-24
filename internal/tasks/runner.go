package tasks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"orchard/internal/project"
	"orchard/internal/tools"
)

// Runner starts tasks and remembers them per repository. Its methods are
// safe from any goroutine; changed is called (at most every 50ms) when
// a task's output or state changes, so a UI can redraw.
type Runner struct {
	mu      sync.Mutex
	nextID  int
	tasks   []*Task
	changed func()
	pending bool
}

// keepFinished is how many finished tasks a repository keeps.
const keepFinished = 6

// NewRunner returns a runner reporting changes to changed.
func NewRunner(changed func()) *Runner { return &Runner{changed: changed} }

// Start runs steps one after another in dir, stopping at the first
// failure. A running task of the same repository and kind is stopped
// first, so pressing Run again restarts the app.
func (r *Runner) Start(repoID, dir, kind, title string, steps []project.Command) *Task {
	r.mu.Lock()
	for _, t := range r.tasks {
		if t.RepoID == repoID && t.Kind == kind && t.State() == Running {
			defer t.Stop()
		}
	}
	r.nextID++
	t := &Task{ID: r.nextID, RepoID: repoID, Title: title, Kind: kind, Dir: dir, Steps: steps, Started: time.Now()}
	r.tasks = append(r.tasks, t)
	r.prune(repoID)
	r.mu.Unlock()
	go r.run(t)
	r.notify()
	return t
}

// Tasks returns a repository's tasks, oldest first.
func (r *Runner) Tasks(repoID string) []*Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*Task
	for _, t := range r.tasks {
		if t.RepoID == repoID {
			out = append(out, t)
		}
	}
	return out
}

// Running reports whether a repository has a running task of a kind.
func (r *Runner) Running(repoID, kind string) *Task {
	for _, t := range r.Tasks(repoID) {
		if t.Kind == kind && t.State() == Running {
			return t
		}
	}
	return nil
}

// Remove forgets a finished task.
func (r *Runner) Remove(t *Task) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, x := range r.tasks {
		if x == t && t.State() != Running {
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
			break
		}
	}
	r.notify()
}

// StopAll stops every running task, when the app quits.
func (r *Runner) StopAll() {
	r.mu.Lock()
	tasks := append([]*Task(nil), r.tasks...)
	r.mu.Unlock()
	for _, t := range tasks {
		if t.State() == Running {
			t.Stop()
		}
	}
}

// prune drops a repository's oldest finished tasks beyond keepFinished.
func (r *Runner) prune(repoID string) {
	finished := 0
	for i := len(r.tasks) - 1; i >= 0; i-- {
		t := r.tasks[i]
		if t.RepoID != repoID || t.State() == Running {
			continue
		}
		if finished++; finished > keepFinished {
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
		}
	}
}

func (r *Runner) run(t *Task) {
	var err error
	for _, step := range t.Steps {
		if err = r.runStep(t, step); err != nil {
			break
		}
		t.mu.Lock()
		stopped := t.stopped
		t.mu.Unlock()
		if stopped {
			break
		}
	}
	t.finish(err)
	t.append(Line{Text: finishLine(t), Info: true})
	r.notify()
}

func finishLine(t *Task) string {
	s := t.Snapshot()
	d := s.Ended.Sub(t.Started).Round(100 * time.Millisecond)
	switch s.State {
	case Succeeded:
		return fmt.Sprintf("Finished in %s", d)
	case Stopped:
		return fmt.Sprintf("Stopped after %s", d)
	}
	return fmt.Sprintf("Failed after %s: %v", d, s.Err)
}

func (r *Runner) runStep(t *Task, step project.Command) error {
	t.append(Line{Text: "$ " + strings.Join(step.Args, " "), Info: true})
	r.notify()
	bin, err := tools.Find(step.Args[0])
	if err != nil {
		if hint := project.InstallHint(step.Args[0]); hint != "" {
			t.append(Line{Text: hint, Info: true})
		}
		return err
	}
	cmd := exec.CommandContext(context.Background(), bin, step.Args[1:]...)
	cmd.Dir = t.Dir
	cmd.Env = append(os.Environ(), "PATH="+tools.SearchPath(), "FORCE_COLOR=0", "NO_COLOR=1")
	cmd.Stdin = nil
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	isolate(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	t.mu.Lock()
	t.stop = func() { killTree(cmd) }
	stopped := t.stopped
	t.mu.Unlock()
	if stopped {
		killTree(cmd)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go r.pump(t, stdout, false, &wg)
	go r.pump(t, stderr, true, &wg)
	wg.Wait()
	return cmd.Wait()
}

// pump copies a pipe into the task's output, a line at a time; progress
// redrawn with carriage returns keeps only its latest state.
func (r *Runner) pump(t *Task, rd io.Reader, stderr bool, wg *sync.WaitGroup) {
	defer wg.Done()
	sc := bufio.NewScanner(rd)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		text := sc.Text()
		if i := strings.LastIndex(strings.TrimRight(text, "\r"), "\r"); i >= 0 {
			text = text[i+1:]
		}
		t.append(Line{Text: stripANSI(text), Stderr: stderr})
		r.notify()
	}
}

// notify tells the UI about changes, at most every 50ms.
func (r *Runner) notify() {
	r.mu.Lock()
	if r.pending || r.changed == nil {
		r.mu.Unlock()
		return
	}
	r.pending = true
	r.mu.Unlock()
	time.AfterFunc(50*time.Millisecond, func() {
		r.mu.Lock()
		r.pending = false
		r.mu.Unlock()
		r.changed()
	})
}
