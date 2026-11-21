package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

const testHeartbeatToken = "s3cr3t-heartbeat-token"

// captureLogs redirects the package logger into a buffer at debug level for the duration of the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	oldLevel := logger.GetLevel()
	logger.SetOutput(buf)
	logger.SetLevel(logrus.DebugLevel)

	t.Cleanup(func() {
		logger.SetOutput(os.Stderr)
		logger.SetLevel(oldLevel)
	})

	return buf
}

func TestEmitHeartbeat_SendsTokenToTarget(t *testing.T) {
	var mu sync.Mutex
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotPath = r.URL.Path
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("HEARTBEAT_TOKEN", testHeartbeatToken)
	t.Setenv("HEARTBEAT_TARGET", srv.URL+"/api/v1/heartbeat/")
	captureLogs(t)

	emitHeartbeat()

	mu.Lock()
	defer mu.Unlock()
	if want := "/api/v1/heartbeat/" + testHeartbeatToken; gotPath != want {
		t.Errorf("heartbeat requested path %q, want %q", gotPath, want)
	}
}

func TestEmitHeartbeat_NeverLogsTheToken(t *testing.T) {
	t.Run("connection error", func(t *testing.T) {
		// A closed server makes client.Get fail with a *url.Error that embeds the full request URL.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		target := srv.URL + "/api/v1/heartbeat/"
		srv.Close()

		t.Setenv("HEARTBEAT_TOKEN", testHeartbeatToken)
		t.Setenv("HEARTBEAT_TARGET", target)
		logs := captureLogs(t)

		emitHeartbeat()

		out := logs.String()
		if strings.Contains(out, testHeartbeatToken) {
			t.Errorf("log output contains the heartbeat token:\n%s", out)
		}
		if !strings.Contains(out, "Unable to emit heartbeat") {
			t.Errorf("expected the failure to be logged, got:\n%s", out)
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		t.Setenv("HEARTBEAT_TOKEN", testHeartbeatToken)
		t.Setenv("HEARTBEAT_TARGET", srv.URL+"/api/v1/heartbeat/")
		logs := captureLogs(t)

		emitHeartbeat()

		out := logs.String()
		if strings.Contains(out, testHeartbeatToken) {
			t.Errorf("log output contains the heartbeat token:\n%s", out)
		}
		if !strings.Contains(out, "Emitting heartbeat failed") {
			t.Errorf("expected the failure to be logged, got:\n%s", out)
		}
	})

	t.Run("success path at debug level", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		t.Setenv("HEARTBEAT_TOKEN", testHeartbeatToken)
		t.Setenv("HEARTBEAT_TARGET", srv.URL+"/api/v1/heartbeat/")
		logs := captureLogs(t)

		emitHeartbeat()

		if out := logs.String(); strings.Contains(out, testHeartbeatToken) {
			t.Errorf("debug log output contains the heartbeat token:\n%s", out)
		}
	})
}

func TestEmitHeartbeat_NoTokenIsANoop(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	t.Setenv("HEARTBEAT_TOKEN", "")
	t.Setenv("HEARTBEAT_TARGET", srv.URL+"/api/v1/heartbeat/")
	captureLogs(t)

	emitHeartbeat()

	if called {
		t.Error("no heartbeat should be sent without a token")
	}
}
