package gitx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// CloneOptions configures Clone.
type CloneOptions struct {
	URL       string
	Dest      string
	Branch    string
	Depth     int
	Recursive bool
	// AuthUser and AuthToken, when set, authenticate an HTTPS clone for
	// this one command without writing the token into .git/config.
	AuthUser  string
	AuthToken string
}

// Progress is a step of a long-running git operation.
type Progress struct {
	Phase   string
	Percent int
}

var progressLine = regexp.MustCompile(`^([A-Za-z ]+):\s+(\d+)%`)

// Clone clones a repository, reporting progress as git prints it.
func Clone(ctx context.Context, opts CloneOptions, onProgress func(Progress)) error {
	if err := os.MkdirAll(filepath.Dir(opts.Dest), 0o755); err != nil {
		return err
	}
	var args []string
	if opts.AuthToken != "" {
		cred := base64.StdEncoding.EncodeToString([]byte(opts.AuthUser + ":" + opts.AuthToken))
		args = append(args, "-c", "http.extraHeader=Authorization: Basic "+cred)
	}
	args = append(args, "clone", "--progress")
	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}
	if opts.Depth > 0 {
		args = append(args, "--depth", strconv.Itoa(opts.Depth))
	}
	if opts.Recursive {
		args = append(args, "--recurse-submodules")
	}
	args = append(args, "--", opts.URL, opts.Dest)

	cmd := command(ctx, "", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	var tail bytes.Buffer
	scanner := bufio.NewScanner(stderr)
	scanner.Split(splitProgress)
	for scanner.Scan() {
		line := scanner.Text()
		tail.WriteString(line + "\n")
		if m := progressLine.FindStringSubmatch(strings.TrimPrefix(line, "remote: ")); m != nil && onProgress != nil {
			pct, _ := strconv.Atoi(m[2])
			onProgress(Progress{Phase: strings.TrimSpace(m[1]), Percent: pct})
		}
	}
	if err := cmd.Wait(); err != nil {
		return &Error{Args: args, Stderr: tail.String(), Err: err}
	}
	return nil
}

// splitProgress splits git's progress output on both \r and \n, since
// git redraws progress lines in place with carriage returns.
func splitProgress(data []byte, atEOF bool) (int, []byte, error) {
	for i, b := range data {
		if b == '\r' || b == '\n' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// RepoNameFromURL guesses the folder name git would clone url into.
func RepoNameFromURL(url string) string {
	url = strings.TrimRight(strings.TrimSpace(url), "/")
	url = strings.TrimSuffix(url, ".git")
	if i := strings.LastIndexAny(url, "/:"); i >= 0 {
		url = url[i+1:]
	}
	return url
}
