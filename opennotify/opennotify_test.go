package opennotify_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/opennotify-cli/opennotify"
)

const fakeISSJSON = `{
  "message": "success",
  "iss_position": {
    "latitude": "31.7093",
    "longitude": "20.3513"
  },
  "timestamp": 1781438581
}`

const fakeAstrosJSON = `{
  "message": "success",
  "number": 3,
  "people": [
    {"name": "Oleg Kononenko", "craft": "ISS"},
    {"name": "Sunita Williams", "craft": "ISS"},
    {"name": "Li Guangsu",     "craft": "Tiangong"}
  ]
}`

const fakeErrorJSON = `{"message": "failure", "iss_position": {"latitude": "0", "longitude": "0"}, "timestamp": 0}`

func newTestClient(ts *httptest.Server) *opennotify.Client {
	cfg := opennotify.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return opennotify.NewClient(cfg)
}

// TestPositionSendsUserAgent verifies that every position request
// includes a non-empty User-Agent header.
func TestPositionSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeISSJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Position(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

// TestPositionParses verifies that string lat/lon fields are decoded correctly
// and the timestamp is set.
func TestPositionParses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeISSJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	pos, err := c.Position(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if pos.Latitude != "31.7093" {
		t.Errorf("Latitude = %q, want %q", pos.Latitude, "31.7093")
	}
	if pos.Longitude != "20.3513" {
		t.Errorf("Longitude = %q, want %q", pos.Longitude, "20.3513")
	}
	if pos.Timestamp != 1781438581 {
		t.Errorf("Timestamp = %d, want 1781438581", pos.Timestamp)
	}
}

// TestPositionRetriesOn503 verifies that the client retries on HTTP 503
// and eventually succeeds when the server recovers.
func TestPositionRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeISSJSON)
	}))
	defer ts.Close()

	cfg := opennotify.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := opennotify.NewClient(cfg)

	_, err := c.Position(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

// TestAstronautsParses verifies that the people list is decoded correctly.
func TestAstronautsParses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeAstrosJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	people, err := c.Astronauts(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(people) != 3 {
		t.Fatalf("len(people) = %d, want 3", len(people))
	}
	if people[0].Name != "Oleg Kononenko" {
		t.Errorf("people[0].Name = %q, want \"Oleg Kononenko\"", people[0].Name)
	}
	if people[0].Craft != "ISS" {
		t.Errorf("people[0].Craft = %q, want \"ISS\"", people[0].Craft)
	}
	if people[2].Craft != "Tiangong" {
		t.Errorf("people[2].Craft = %q, want \"Tiangong\"", people[2].Craft)
	}
}

// TestPositionAPIError verifies that a non-success message returns an error.
func TestPositionAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeErrorJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Position(context.Background())
	if err == nil {
		t.Error("expected error for non-success API response, got nil")
	}
}
