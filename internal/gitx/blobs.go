package gitx

import (
	"sort"
	"strconv"
	"strings"
)

// BlobInfo is a file version stored in the repository.
type BlobInfo struct {
	Path string
	Size int64
}

// topBlobs keeps the largest version of each path from cat-file output.
func topBlobs(out string, limit int) []BlobInfo {
	largest := map[string]int64{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.SplitN(line, " ", 3)
		if len(f) < 3 || f[0] != "blob" || f[2] == "" {
			continue
		}
		size, err := strconv.ParseInt(f[1], 10, 64)
		if err == nil && size > largest[f[2]] {
			largest[f[2]] = size
		}
	}
	blobs := make([]BlobInfo, 0, len(largest))
	for p, s := range largest {
		blobs = append(blobs, BlobInfo{Path: p, Size: s})
	}
	sort.Slice(blobs, func(i, j int) bool { return blobs[i].Size > blobs[j].Size })
	if limit > 0 && len(blobs) > limit {
		blobs = blobs[:limit]
	}
	return blobs
}
