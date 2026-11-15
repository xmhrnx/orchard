package gitx

import (
	"strconv"
	"strings"
	"time"
)

func parseUnix(s string) time.Time {
	secs, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}
