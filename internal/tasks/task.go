// Package tasks runs project commands (run, build, dev servers) as
// child processes, keeps their output, and stops them on request.
package tasks

import (
	"sync"
	"time"

	"orchard/internal/project"
)

// State is where a task is in its life.
type State int

// Task states.
const (
	Running State = iota
	Succeeded
	Failed
	Stopped
)

// maxLines bounds the output a task keeps; older lines are dropped.
const maxLines = 5000

// Line is one line of a task's output.
type Line struct {
	Text   string
	Stderr bool
	// Info marks lines Orchard writes itself, such as each command.
	Info bool
}

// Task is one run of a sequence of commands in a repository.
type Task struct {
	ID     int
	RepoID string
	Title  string
	// Kind groups tasks by what they do ("run", "build", ...), so a
	// second start of the same kind can replace the first.
	Kind    string
	Dir     string
	Steps   []project.Command
	Started time.Time

	mu      sync.Mutex
	state   State
	ended   time.Time
	err     error
	lines   []Line
	dropped int
	stop    func()
	stopped bool
}

// Snapshot is a consistent copy of a task's changing state.
type Snapshot struct {
	State   State
	Ended   time.Time
	Err     error
	Lines   []Line
	Dropped int
}

// Snapshot copies the task's state and output for display.
func (t *Task) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	return Snapshot{State: t.state, Ended: t.ended, Err: t.err, Lines: append([]Line(nil), t.lines...), Dropped: t.dropped}
}

// State is the task's current state.
func (t *Task) State() State {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}

// Stop ends the task and every process it started.
func (t *Task) Stop() {
	t.mu.Lock()
	stop := t.stop
	t.stopped = true
	t.mu.Unlock()
	if stop != nil {
		stop()
	}
}

func (t *Task) append(l Line) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lines = append(t.lines, l)
	if over := len(t.lines) - maxLines; over > 0 {
		t.lines = append([]Line(nil), t.lines[over:]...)
		t.dropped += over
	}
}

func (t *Task) finish(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ended, t.err, t.stop = time.Now(), err, nil
	switch {
	case t.stopped:
		t.state = Stopped
	case err != nil:
		t.state = Failed
	default:
		t.state = Succeeded
	}
}

// Title is the state as shown to people.
func (s State) Title() string {
	return [...]string{"Running", "Succeeded", "Failed", "Stopped"}[s]
}
