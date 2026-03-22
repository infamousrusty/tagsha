// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

//go:build integration

// Integration tests require a running stack.
// Run: TAGSHA_API_URL=http://localhost:8080 go test -tags integration ./tests/integration/...

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var apiURL string

func init() {
	apiURL = os.Getenv("TAGSHA_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
}

func TestHealth_Integration(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", body["status"])
	}
}

func TestResolveThenTags_Integration(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}
	body := strings.NewReader(`{"query":"golang/go"}`)
	resp, err := client.Post(apiURL+"/api/v1/resolve", "application/json", body)
	if err != nil {
		t.Fatalf("resolve request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("resolve status = %d, want 200", resp.StatusCode)
	}
	var resolved struct {
		Owner string `json:"owner"`
		Repo  string `json:"repo"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&resolved); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resolved.Owner != "golang" {
		t.Errorf("owner = %s, want golang", resolved.Owner)
	}

	tags, err := client.Get(fmt.Sprintf("%s/api/v1/tags/%s/%s", apiURL, resolved.Owner, resolved.Repo))
	if err != nil {
		t.Fatalf("tags request failed: %v", err)
	}
	defer tags.Body.Close()
	if tags.StatusCode != 200 {
		t.Errorf("tags status = %d, want 200", tags.StatusCode)
	}
}

func TestSSRFRejection_Integration(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}
	badInputs := []string{
		`{"query":"https://evil.com/owner/repo"}`,
		`{"query":"https://10.0.0.1/owner/repo"}`,
		`{"query":""}`,
		`{"query":"../../../etc/passwd"}`,
	}
	for _, input := range badInputs {
		resp, err := client.Post(apiURL+"/api/v1/resolve", "application/json", strings.NewReader(input))
		if err != nil {
			t.Errorf("request failed for %s: %v", input, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 400 {
			t.Errorf("SSRF input %s returned %d, want 400", input, resp.StatusCode)
		}
	}
}

func TestSecurityHeaders_Integration(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	}
	for k, v := range want {
		if got := resp.Header.Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}
