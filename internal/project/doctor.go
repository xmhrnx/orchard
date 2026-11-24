package project

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"orchard/internal/tools"
)

// RunDoctor runs giode doctor in a project folder and reads its report.
// It runs there because Go picks the toolchain the project's go.mod
// asks for, as the build will. Doctor exits with an error when it finds
// problems, so only a doctor that couldn't run at all is an error.
func RunDoctor(ctx context.Context, dir string) (Doctor, error) {
	bin, err := tools.Find("giode")
	if err != nil {
		return Doctor{}, err
	}
	cmd := exec.CommandContext(ctx, bin, "doctor")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+tools.SearchPath(), "NO_COLOR=1")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	runErr := cmd.Run()
	d := ParseDoctor(out.String())
	if len(d.Sections) == 0 && runErr != nil {
		return d, runErr
	}
	return d, nil
}
