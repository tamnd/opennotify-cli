package opennotify

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string
// functions. No network is needed.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "opennotify" {
		t.Errorf("Scheme = %q, want opennotify", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "opennotify" {
		t.Errorf("Identity.Binary = %q, want opennotify", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	typ, id, err := Domain{}.Classify("iss-now")
	if err != nil {
		t.Fatal(err)
	}
	if typ != "position" {
		t.Errorf("Classify type = %q, want \"position\"", typ)
	}
	if id == "" {
		t.Error("Classify id is empty")
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("position", "any")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Error("Locate returned empty URL")
	}
}

func TestLocateAstronauts(t *testing.T) {
	got, err := Domain{}.Locate("astronauts", "any")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Error("Locate returned empty URL for astronauts")
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}
