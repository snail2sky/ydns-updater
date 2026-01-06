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

	var (
		host     string
		ip       string
		recordID string
	)

	// Setup the testing server to get the request.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the passed "host" request parameter as the external "host" var.
		host = r.URL.Query().Get("host")
		ip = r.URL.Query().Get("ip")
		recordID = r.URL.Query().Get("record_id")
		auth := r.Header.Get("authorization")
		fmt.Println(auth)
		r.URL.User.Username()

		// Send the OK back.
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	requestInfo := ydns.RequestInfo{
		Base:     ts.URL,
		Host:     "test-1.com",
		IP:       "127.0.0.1",
		RecordID: "123456",
		User:     "user",
		Pass:     "pass",
	}

	// Ensure there's no error running it with the right options.
	if err := ydns.Run(&requestInfo); err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}

	if host != "test-1.com" {
		t.Fatalf("expected host to equal test-1.com, got: %s", host)
	}

	if ip != "127.0.0.1" {
		t.Fatalf("expected ip to equal x.x.x.x, got: %s", ip)
	}

	if recordID != "123456" {
		t.Fatalf("expected record_id to equal record_id, got: %s", recordID)
	}
}
