package core

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"orchard/internal/forge"
)

// Accounts returns the connected hosting accounts.
func (w *Workspace) Accounts() []forge.Account { return w.state.Accounts }

// AddAccount verifies a token against the service and, when it works,
// saves the account with the token in the vault. done runs on the UI
// goroutine.
func (w *Workspace) AddAccount(kind forge.Kind, host, username, token string, done func(error)) {
	acc := forge.Account{
		ID:       fmt.Sprintf("%s-%d", kind, time.Now().UnixNano()),
		Kind:     kind,
		Host:     strings.ToLower(strings.TrimSpace(host)),
		Username: strings.TrimSpace(username),
	}
	if acc.Host == "" {
		acc.Host = forge.DefaultHost(kind)
	}
	go func() {
		client, err := forge.New(acc, token)
		if err == nil {
			var login string
			login, err = client.CurrentUser(w.ctx)
			if err == nil && login != "" {
				acc.Username = login
			}
		}
		if err == nil {
			err = w.vault.Set(acc.TokenKey(), []byte(token))
		}
		w.post(func() {
			if err == nil {
				w.state.Accounts = append(w.state.Accounts, acc)
				err = w.save()
			}
			done(err)
		})
	}()
}

// RemoveAccount forgets an account and deletes its token.
func (w *Workspace) RemoveAccount(id string) error {
	var kept []forge.Account
	var removed *forge.Account
	for _, a := range w.state.Accounts {
		if a.ID == id {
			removed = &a
			continue
		}
		kept = append(kept, a)
	}
	w.state.Accounts = kept
	err := w.save()
	if removed != nil {
		err = errors.Join(err, w.vault.Delete(removed.TokenKey()))
	}
	return err
}

// Client builds the API client of an account. It reads the vault, which
// may block, so call it off the UI goroutine.
func (w *Workspace) Client(acc forge.Account) (forge.Client, error) {
	token, err := w.vault.Get(acc.TokenKey())
	if err != nil {
		return nil, fmt.Errorf("no token for %s: %w", acc.Host, err)
	}
	return forge.New(acc, string(token))
}

// HostedRepo finds the account and "owner/name" of a repository's
// origin, for pull request features.
func (w *Workspace) HostedRepo(r *RepoState) (forge.Account, string, bool) {
	loc, ok := forge.ParseRemote(r.Origin)
	if !ok {
		return forge.Account{}, "", false
	}
	acc, ok := forge.MatchAccount(w.state.Accounts, loc)
	return acc, loc.Repo, ok
}
