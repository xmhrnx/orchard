package forge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSortReleasesMarksTheNewestStableAsLatest(t *testing.T) {
	now := time.Now()
	list := sortReleases([]Release{
		{Tag: "v1", Published: now.Add(-48 * time.Hour)},
		{Tag: "v3-rc", Prerelease: true, Published: now},
		{Tag: "v2", Published: now.Add(-time.Hour)},
		{Tag: "v4", Draft: true},
	})
	order := ""
	for _, r := range list {
		order += r.Tag + " "
	}
	if order != "v4 v3-rc v2 v1 " {
		t.Errorf("order = %q, want drafts first then newest", order)
	}
	for _, r := range list {
		if r.Latest != (r.Tag == "v2") {
			t.Errorf("%s latest = %v", r.Tag, r.Latest)
		}
	}
}

func TestGithubReleases(t *testing.T) {
	var created map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode([]map[string]any{{
				"id": 5, "tag_name": "v1.0.0", "name": "First", "body": "notes",
				"published_at": time.Now().Format(time.RFC3339), "author": map[string]string{"login": "dev"},
				"assets": []map[string]any{{"name": "app.zip", "size": 2048, "download_count": 9}},
			}})
		case http.MethodPost:
			json.NewDecoder(r.Body).Decode(&created)
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer srv.Close()
	c := &githubClient{http: newHTTP("t"), base: srv.URL}
	list, err := c.Releases(context.Background(), "o/r")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || !list[0].Latest || list[0].Assets[0].Downloads != 9 {
		t.Fatalf("releases = %+v", list)
	}
	err = c.CreateRelease(context.Background(), "o/r", NewRelease{Tag: "v2", Target: "main", Draft: true, GenerateNotes: true})
	if err != nil {
		t.Fatal(err)
	}
	if created["tag_name"] != "v2" || created["target_commitish"] != "main" || created["draft"] != true || created["generate_release_notes"] != true {
		t.Errorf("create body = %v", created)
	}
}

func TestEveryClientExceptBitbucketHasReleases(t *testing.T) {
	for _, k := range Kinds {
		c, err := New(Account{Kind: k, Host: DefaultHost(k)}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := ReleasesOf(c); ok == (k == Bitbucket) {
			t.Errorf("%s: releases supported = %v", k, ok)
		}
	}
}
