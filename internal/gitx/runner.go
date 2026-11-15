// Package gitx drives the git command-line tool. Every operation runs git
// as a child process, so it works with any host that speaks git: GitHub,
// GitLab, Gitea, Bitbucket, a bare repository on a file share, ...
package gitx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Error is a failed git invocation with the output git printed.
type Error struct {
	Args   []string
	Stderr string
	Err    error
}

// ErrNotARepository is returned when a folder is not inside a git work tree.
var ErrNotARepository = errors.New("not a git repository")

func (e *Error) Error() string {
	msg := strings.TrimSpace(e.Stderr)
	if msg == "" {
		msg = e.Err.Error()
	}
	return fmt.Sprintf("git %s: %s", firstArg(e.Args), lastLine(msg))
}

func (e *Error) Unwrap() error { return e.Err }

// Binary is the git executable to run. It is a variable so tests and
// users with an unusual installation can point it elsewhere.
var Binary = "git"

// run executes git in dir and returns its standard output.
func run(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := runBytes(ctx, dir, nil, args...)
	return string(out), err
}

// runInput executes git with stdin fed from input.
func runInput(ctx context.Context, dir, input string, args ...string) (string, error) {
	out, err := runBytes(ctx, dir, strings.NewReader(input), args...)
	return string(out), err
}

func runBytes(ctx context.Context, dir string, stdin *strings.Reader, args ...string) ([]byte, error) {
	cmd := command(ctx, dir, args...)
	if stdin != nil {
		cmd.Stdin = stdin
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), &Error{Args: args, Stderr: stderr.String(), Err: err}
	}
	return stdout.Bytes(), nil
}

// command builds a git command that never prompts on a terminal: a
// desktop app has none, and a prompt would hang the operation.
func command(ctx context.Context, dir string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, Binary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_OPTIONAL_LOCKS=0",
		"FILTER_BRANCH_SQUELCH_WARNING=1",
		"LC_ALL=C",
	)
	hideWindow(cmd)
	return cmd
}

// Available reports whether the git executable can be run.
func Available() bool {
	_, err := exec.LookPath(Binary)
	return err == nil
}

// Version returns the installed git version, e.g. "2.43.0".
func Version(ctx context.Context) (string, error) {
	out, err := run(ctx, "", "version")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(out), "git version "), nil
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

// lastLine picks the line of git's output that explains a failure: the
// first "fatal:" or "error:" line, else the last line that isn't a hint.
func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "fatal: ") || strings.HasPrefix(l, "error: ") {
			return strings.TrimPrefix(strings.TrimPrefix(l, "fatal: "), "error: ")
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if l := strings.TrimSpace(lines[i]); l != "" && !strings.HasPrefix(l, "hint:") {
			return l
		}
	}
	return s
}
