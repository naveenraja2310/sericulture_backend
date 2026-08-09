package utils

import "testing"

func TestBuildGrafanaBasicAuth(t *testing.T) {
	got := BuildGrafanaBasicAuth("instance-123", "api-key-456")
	want := "aW5zdGFuY2UtMTIzOmFwaS1rZXktNDU2"
	if got != want {
		t.Fatalf("BuildGrafanaBasicAuth() = %q, want %q", got, want)
	}

	if got := BuildGrafanaBasicAuth("", "api-key-456"); got != "" {
		t.Fatalf("BuildGrafanaBasicAuth() should return empty when instanceID is missing")
	}
}
