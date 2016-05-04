package actions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the GitHub client being tested.
	client *github.Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// github client configured to use test server
	client = github.NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
	client.UploadURL = url
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func TestGenerateChangelog(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/deis/controller/compare/b...h", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{
		  "base_commit": {
		    "sha": "s",
		    "commit": {
		      "author": { "name": "n" },
		      "committer": { "name": "n" },
		      "message": "m",
		      "tree": { "sha": "t" }
		    },
		    "author": { "login": "n" },
		    "committer": { "login": "l" },
		    "parents": [ { "sha": "s" } ]
		  },
		  "status": "s",
		  "ahead_by": 1,
		  "behind_by": 2,
		  "total_commits": 1,
		  "commits": [
		    {
		      "sha": "abc1234567890",
		      "commit": { "author": { "name": "n" }, "message": "feat(deisrel): new feature!" },
		      "author": { "login": "l" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
		    },
		    {
		      "sha": "abc2345678901",
		      "commit": { "author": { "name": "n" }, "message": "fix(deisrel): bugfix!" },
		      "author": { "login": "l" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
		    },
		    {
		      "sha": "abc3456789012",
		      "commit": { "author": { "name": "n" }, "message": "docs(deisrel): new docs!" },
		      "author": { "login": "l" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
		    },
		    {
		      "sha": "abc4567890123",
		      "commit": { "author": { "name": "n" }, "message": "chore(deisrel): boring chore" },
		      "author": { "login": "l" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
		    }
		  ],
		  "files": [ { "filename": "f" } ]
		}`)
	})

	got := &Changelog{
		OldRelease: "b",
		NewRelease: "h",
	}

	if err := generateChangelog(client, got); err != nil {
		t.Errorf("generateChangelog returned an error: %s", err)
	}

	want := &Changelog{
		OldRelease: "b",
		NewRelease: "h",
		Features: []string{"abc1234 deisrel: new feature!"},
		Fixes: []string{"abc2345 deisrel: bugfix!"},
		Documentation: []string{"abc3456 deisrel: new docs!"},
		Maintenance: []string{"abc4567 deisrel: boring chore"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("generateChangelog returned \n%+v, want \n%+v", got, want)
	}
}

func TestGenerateChangelog_NoRelevantCommits(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/deis/r/compare/b...h", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{
		  "base_commit": {
		    "sha": "s",
		    "commit": {
		      "author": { "name": "n" },
		      "committer": { "name": "n" },
		      "message": "m",
		      "tree": { "sha": "t" }
		    },
		    "author": { "login": "n" },
		    "committer": { "login": "l" },
		    "parents": [ { "sha": "s" } ]
		  },
		  "status": "s",
		  "ahead_by": 1,
		  "behind_by": 2,
		  "total_commits": 1,
		  "commits": [
		    {
		      "sha": "s",
		      "commit": { "author": { "name": "n" } },
		      "author": { "login": "l" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
		    }
		  ],
		  "files": [ { "filename": "f" } ]
		}`)
	})

	got := &Changelog{
		OldRelease: "b",
		NewRelease: "h",
	}

	if err := generateChangelog(client, got); err != nil {
		t.Errorf("generateChangelog returned an error: %s", err)
	}

	want := &Changelog{
		OldRelease: "b",
		NewRelease: "h",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("generateChangelog returned \n%+v, want \n%+v", got, want)
	}
}
