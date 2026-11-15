package screens

import (
	"context"
	"fmt"

	"orchard/internal/gitx"
)

// createAndPushTag tags ref and, when the repository has an origin,
// pushes the tag there straight away, so a new tag is visible on the
// hosting service and not only on this computer. A name origin already
// uses is refused before anything is created, so no two different tags
// end up sharing it.
func createAndPushTag(ctx context.Context, path, name, ref string) (pushed bool, err error) {
	hasOrigin := gitx.OriginURL(ctx, path) != ""
	if hasOrigin {
		remote, err := gitx.RemoteTags(ctx, path, "origin")
		if err != nil {
			return false, fmt.Errorf("checking origin's tags first: %w", err)
		}
		for _, r := range remote {
			if r.Name == name {
				return false, fmt.Errorf("origin already has a tag named %s (on commit %.7s); fetch tags to get it, or choose another name", name, r.Commit)
			}
		}
	}
	if err := gitx.CreateTag(ctx, path, name, ref, ""); err != nil {
		return false, err
	}
	if !hasOrigin {
		return false, nil
	}
	if err := gitx.PushTag(ctx, path, "origin", name); err != nil {
		return false, fmt.Errorf("tagged %s here, but pushing it to origin failed: %w", name, err)
	}
	return true, nil
}

// tagCreated is the toast for a new tag.
func tagCreated(name string, pushed bool) string {
	if pushed {
		return "Tagged " + name + " and pushed it to origin"
	}
	return "Tagged " + name + " (this repository has no origin to push to)"
}
