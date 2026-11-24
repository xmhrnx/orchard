package screens

import (
	"context"

	"orchard/internal/core"
	"orchard/internal/project"
	"orchard/ui/env"
)

// doctorResult is giode doctor's verdict for one project folder.
type doctorResult struct {
	report  project.Doctor
	err     error
	loaded  bool
	loading bool
}

// doctorCache runs giode doctor once per project and keeps the report.
// Doctor runs in the project because Go uses the toolchain its go.mod
// asks for.
type doctorCache struct {
	env     *env.Env
	results map[string]*doctorResult
}

func newDoctorCache(e *env.Env) *doctorCache {
	return &doctorCache{env: e, results: map[string]*doctorResult{}}
}

// get returns the report for a project folder, starting doctor the
// first time.
func (c *doctorCache) get(dir string) (project.Doctor, bool, error) {
	r := c.results[dir]
	if r == nil {
		r = &doctorResult{}
		c.results[dir] = r
		c.refresh(dir)
	}
	return r.report, r.loaded, r.err
}

// refresh runs doctor again, after installing an SDK for example.
func (c *doctorCache) refresh(dir string) {
	r := c.results[dir]
	if r == nil {
		r = &doctorResult{}
		c.results[dir] = r
	}
	if r.loading {
		return
	}
	r.loading = true
	core.Query(c.env.WS, func(ctx context.Context) (project.Doctor, error) {
		return project.RunDoctor(ctx, dir)
	}, func(d project.Doctor, err error) {
		r.report, r.err, r.loaded, r.loading = d, err, true, false
	})
}
