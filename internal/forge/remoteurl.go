package forge

import (
	"net/url"
	"strings"
)

// RemoteLocation is where a git remote URL points.
type RemoteLocation struct {
	Host string
	Repo string
}

// ParseRemote reads the host and "owner/name" path from any git remote
// URL form: https://host/owner/name.git, git@host:owner/name.git or
// ssh://git@host:22/owner/name.
func ParseRemote(raw string) (RemoteLocation, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RemoteLocation{}, false
	}
	if !strings.Contains(raw, "://") {
		at := strings.Index(raw, "@")
		colon := strings.Index(raw, ":")
		if colon < 0 || colon < at {
			return RemoteLocation{}, false
		}
		return newLocation(raw[at+1:colon], raw[colon+1:])
	}
	u, err := url.Parse(raw)
	if err != nil {
		return RemoteLocation{}, false
	}
	return newLocation(u.Hostname(), u.Path)
}

func newLocation(host, path string) (RemoteLocation, bool) {
	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")
	if host == "" || !strings.Contains(path, "/") {
		return RemoteLocation{}, false
	}
	return RemoteLocation{Host: strings.ToLower(host), Repo: path}, true
}

// MatchAccount returns the account for a remote's host, if any.
func MatchAccount(accounts []Account, loc RemoteLocation) (Account, bool) {
	for _, a := range accounts {
		if strings.EqualFold(a.Host, loc.Host) || (a.Kind == GitHub && loc.Host == "github.com" && a.Host == "") {
			return a, true
		}
	}
	return Account{}, false
}
