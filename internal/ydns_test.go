package ydns_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	ydns "github.com/wyattjoh/ydns-updater/internal"
)

func TestRun(t *testing.T) {
	// This can be run "in parallel".
	t.Parallel()

	var host string
	var ip string
	var record_id string

	// Setup the testing server to get the request.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the passed "host" request parameter as the external "host" var.
		host = r.URL.Query().Get("host")
		ip = r.URL.Query().Get("ip")
		record_id = r.URL.Query().Get("record_id")

		r.URL.User.Username()

		// Send the OK back.
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	// Ensure there's no error running it with the right options.
	if err := ydns.Run(ts.URL, "test-1.com", "x.x.x.x", "record_id", "user", "pass"); err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}

	if host != "test-1.com" {
		t.Fatalf("expected host to equal test-1.com, got: %s", host)
	}

	if ip != "x.x.x.x" {
		t.Fatalf("expected ip to equal x.x.x.x, got: %s", ip)
	}

	if record_id != "record_id" {
		t.Fatalf("expected record_id to equal record_id, got: %s", record_id)
	}
}
