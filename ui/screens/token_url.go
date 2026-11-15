package screens

import "orchard/internal/forge"

// tokenURL is where a service lets you create an access token with the
// permissions Orchard needs.
func tokenURL(k forge.Kind, host string) string {
	if host == "" {
		host = forge.DefaultHost(k)
	}
	switch k {
	case forge.GitHub:
		return "https://" + host + "/settings/tokens/new?scopes=repo,read:org&description=Orchard"
	case forge.GitLab:
		return "https://" + host + "/-/user_settings/personal_access_tokens?name=Orchard&scopes=api"
	case forge.Gitea:
		return "https://" + host + "/user/settings/applications"
	case forge.Bitbucket:
		return "https://bitbucket.org/account/settings/app-passwords/"
	}
	return ""
}

// tokenHint explains what to create on each service.
func tokenHint(k forge.Kind) string {
	switch k {
	case forge.GitHub:
		return "A classic token with the repo scope (or a fine-grained token with Contents and Pull requests access)."
	case forge.GitLab:
		return "A personal access token with the api scope."
	case forge.Gitea:
		return "An access token with repository read/write and user read permissions. Works for Forgejo and Codeberg too."
	case forge.Bitbucket:
		return "An app password with Repositories and Pull requests read/write. Enter your Bitbucket username too."
	}
	return ""
}

// kindTint is a service's accent in the UI.
func kindTint(k forge.Kind) string {
	switch k {
	case forge.GitHub:
		return "theme.text"
	case forge.GitLab:
		return "#FC6D26"
	case forge.Gitea:
		return "#609926"
	case forge.Bitbucket:
		return "#2684FF"
	}
	return "theme.primary"
}
