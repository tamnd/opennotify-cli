// Package opennotify is the library behind the opennotify command line:
// the HTTP client, request shaping, and the typed data models for the
// Open Notify API (api.open-notify.org).
//
// The API is completely open: no authentication, no API key. It exposes
// two endpoints — the current ISS position and the list of people in space.
// The API is HTTP-only; HTTPS is not supported.
package opennotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Host is the API hostname.
const Host = "api.open-notify.org"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration // minimum gap between consecutive HTTP requests
	Timeout   time.Duration // per-request HTTP timeout
	Retries   int           // maximum retry attempts on transient errors
}

// DefaultConfig returns a Config with sensible defaults for the Open Notify API.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "http://api.open-notify.org",
		UserAgent: "opennotify-cli/0.1 (tamnd87@gmail.com)",
		Rate:      500 * time.Millisecond,
		Timeout:   10 * time.Second,
		Retries:   3,
	}
}

// Client talks to the Open Notify API over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Position returns the current position of the International Space Station.
func (c *Client) Position(ctx context.Context) (*Position, error) {
	u := c.cfg.BaseURL + "/iss-now.json"
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp wirePosition
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode iss-now: %w", err)
	}
	if resp.Message != "success" {
		return nil, fmt.Errorf("api error: %s", resp.Message)
	}
	return &Position{
		Latitude:  resp.ISSPosition.Latitude,
		Longitude: resp.ISSPosition.Longitude,
		Timestamp: resp.Timestamp,
	}, nil
}

// Astronauts returns all humans currently in space.
func (c *Client) Astronauts(ctx context.Context) ([]Astronaut, error) {
	u := c.cfg.BaseURL + "/astros.json"
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp wireAstros
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode astros: %w", err)
	}
	if resp.Message != "success" {
		return nil, fmt.Errorf("api error: %s", resp.Message)
	}
	out := make([]Astronaut, 0, len(resp.People))
	for _, p := range resp.People {
		out = append(out, Astronaut{
			Name:  p.Name,
			Craft: p.Craft,
		})
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
