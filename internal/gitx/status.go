package gitx

import (
	"context"
	"strconv"
	"strings"
)

// ChangeKind classifies a changed file.
type ChangeKind int

// The kinds of change a file can have.
const (
	Modified ChangeKind = iota
	Added
	Deleted
	Renamed
	Copied
	Untracked
	Conflicted
	TypeChanged
)

// FileChange is one changed path in the work tree or the index.
type FileChange struct {
	Path     string
	OrigPath string
	Kind     ChangeKind
}

// Status is a snapshot of a repository's branch and working state.
type Status struct {
	Branch     string
	Upstream   string
	Head       string
	Detached   bool
	Ahead      int
	Behind     int
	Stashes    int
	Staged     []FileChange
	Unstaged   []FileChange
	Untracked  []FileChange
	Conflicted []FileChange
}

// Dirty reports whether anything is changed, staged or untracked.
func (s Status) Dirty() bool {
	return len(s.Staged)+len(s.Unstaged)+len(s.Untracked)+len(s.Conflicted) > 0
}

// ChangeCount is the number of distinct changed paths.
func (s Status) ChangeCount() int {
	seen := map[string]bool{}
	for _, list := range [][]FileChange{s.Staged, s.Unstaged, s.Untracked, s.Conflicted} {
		for _, c := range list {
			seen[c.Path] = true
		}
	}
	return len(seen)
}

// ReadStatus returns the status of the repository at dir.
func ReadStatus(ctx context.Context, dir string) (Status, error) {
	out, err := run(ctx, dir, "status", "--porcelain=v2", "--branch", "--show-stash", "-z", "--untracked-files=all")
	if err != nil {
		return Status{}, err
	}
	return parseStatus(out), nil
}

func parseStatus(out string) Status {
	var st Status
	fields := strings.Split(out, "\x00")
	for i := 0; i < len(fields); i++ {
		line := fields[i]
		if line == "" {
			continue
		}
		switch line[0] {
		case '#':
			parseStatusHeader(&st, line)
		case '1':
			parseOrdinary(&st, line, 8, "")
		case '2':
			orig := ""
			if i+1 < len(fields) {
				orig = fields[i+1]
				i++
			}
			parseOrdinary(&st, line, 9, orig)
		case 'u':
			parts := strings.SplitN(line, " ", 11)
			if len(parts) == 11 {
				st.Conflicted = append(st.Conflicted, FileChange{Path: parts[10], Kind: Conflicted})
			}
		case '?':
			st.Untracked = append(st.Untracked, FileChange{Path: line[2:], Kind: Untracked})
		}
	}
	return st
}

func parseStatusHeader(st *Status, line string) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return
	}
	switch parts[1] {
	case "branch.oid":
		st.Head = parts[2]
	case "branch.head":
		if parts[2] == "(detached)" {
			st.Detached = true
		} else {
			st.Branch = parts[2]
		}
	case "branch.upstream":
		st.Upstream = parts[2]
	case "branch.ab":
		if len(parts) >= 4 {
			st.Ahead, _ = strconv.Atoi(strings.TrimPrefix(parts[2], "+"))
			st.Behind, _ = strconv.Atoi(strings.TrimPrefix(parts[3], "-"))
		}
	case "stash":
		st.Stashes, _ = strconv.Atoi(parts[2])
	}
}

// parseOrdinary reads a "1" or "2" entry: the path is field index
// pathField once split on spaces.
func parseOrdinary(st *Status, line string, pathField int, orig string) {
	parts := strings.SplitN(line, " ", pathField+1)
	if len(parts) <= pathField {
		return
	}
	xy, path := parts[1], parts[pathField]
	if xy[0] != '.' {
		st.Staged = append(st.Staged, FileChange{Path: path, OrigPath: orig, Kind: kindOf(xy[0])})
	}
	if xy[1] != '.' {
		st.Unstaged = append(st.Unstaged, FileChange{Path: path, OrigPath: orig, Kind: kindOf(xy[1])})
	}
}

func kindOf(c byte) ChangeKind {
	switch c {
	case 'A':
		return Added
	case 'D':
		return Deleted
	case 'R':
		return Renamed
	case 'C':
		return Copied
	case 'T':
		return TypeChanged
	case 'U':
		return Conflicted
	default:
		return Modified
	}
}

// Letter is the one-letter code shown next to a changed file.
func (k ChangeKind) Letter() string {
	return [...]string{"M", "A", "D", "R", "C", "U", "!", "T"}[k]
}
