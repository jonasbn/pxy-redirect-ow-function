package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeartbeatTimeout(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"unset uses the default", "", defaultHeartbeatTimeout},
		{"positive value is used", "5", 5 * time.Second},
		{"one second is used", "1", 1 * time.Second},
		// http.Client treats a zero Timeout as "no timeout", which would let a hung heartbeat
		// endpoint hold up every request: non-positive values must fall back to the default.
		{"zero falls back to the default", "0", defaultHeartbeatTimeout},
		{"negative falls back to the default", "-3", defaultHeartbeatTimeout},
		{"not a number falls back to the default", "abc", defaultHeartbeatTimeout},
		{"fractional value falls back to the default", "1.5", defaultHeartbeatTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HEARTBEAT_TARGET_TIMEOUT", tt.value)

			if got := heartbeatTimeout(); got != tt.want {
				t.Errorf("heartbeatTimeout() with %q = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestHeartbeatTimeout_DefaultIsShort(t *testing.T) {
	// The heartbeat is sent before the redirect is returned, so the default is the worst case latency it adds.
	if defaultHeartbeatTimeout > 5*time.Second {
		t.Errorf("defaultHeartbeatTimeout = %v, want at most 5s", defaultHeartbeatTimeout)
	}
}

func TestEmitHeartbeat_ReturnsAfterTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release // a heartbeat endpoint that never answers
	}))
	defer srv.Close()
	defer close(release)

	t.Setenv("HEARTBEAT_TOKEN", "token")
	t.Setenv("HEARTBEAT_TARGET", srv.URL+"/api/v1/heartbeat/")
	t.Setenv("HEARTBEAT_TARGET_TIMEOUT", "1")

	start := time.Now()
	emitHeartbeat()
	elapsed := time.Since(start)

	if elapsed < 900*time.Millisecond {
		t.Errorf("emitHeartbeat returned after %v, before the 1s timeout", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("emitHeartbeat took %v with a 1s timeout", elapsed)
	}
}
