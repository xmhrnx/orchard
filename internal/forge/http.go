package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// httpClient is a small JSON client shared by the service clients.
type httpClient struct {
	token     string
	basicUser string
	client    *http.Client
	header    func(*http.Request)
}

func newHTTP(token string) *httpClient {
	return &httpClient{token: token, client: &http.Client{Timeout: 30 * time.Second}}
}

// withBasic switches to HTTP basic auth (Bitbucket app passwords).
func (h *httpClient) withBasic(user string) *httpClient {
	c := *h
	c.basicUser = user
	return &c
}

// do sends a request and decodes a JSON reply into out (when non-nil).
func (h *httpClient) do(ctx context.Context, method, url string, body any, out any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(data))
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Orchard")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	switch {
	case h.basicUser != "":
		req.SetBasicAuth(h.basicUser, h.token)
	case h.token != "":
		req.Header.Set("Authorization", "Bearer "+h.token)
	}
	if h.header != nil {
		h.header(req)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return resp, fmt.Errorf("%s %s: %s", method, resp.Request.URL.Path, apiMessage(resp.Status, msg))
	}
	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp, fmt.Errorf("decoding %s: %w", resp.Request.URL.Path, err)
		}
	}
	return resp, nil
}

func (h *httpClient) get(ctx context.Context, url string, out any) (*http.Response, error) {
	return h.do(ctx, http.MethodGet, url, nil, out)
}

// apiMessage extracts the "message" most APIs put in an error reply.
func apiMessage(status string, body []byte) string {
	var m struct {
		Message string `json:"message"`
		Error   any    `json:"error"`
	}
	if json.Unmarshal(body, &m) == nil && m.Message != "" {
		return m.Message
	}
	return status
}

// maxPages bounds pagination so a huge organisation can't stall the UI.
const maxPages = 20
