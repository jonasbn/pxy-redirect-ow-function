package main

import (
	"net/url"
	"strings"
	"testing"
)

func mustParseRedirectURL(t *testing.T, path string) *url.URL {
	t.Helper()
	u, err := parseRedirectURL(path)
	if err != nil {
		t.Fatalf("parseRedirectURL(%q) returned unexpected error: %v", path, err)
	}
	return u
}

func TestParseRedirectURL(t *testing.T) {
	t.Run("valid path", func(t *testing.T) {
		u, err := parseRedirectURL("/13/wall")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Path != "/13/wall" {
			t.Errorf("got path %q, want %q", u.Path, "/13/wall")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		if _, err := parseRedirectURL(""); err != nil {
			t.Errorf("unexpected error for empty path: %v", err)
		}
	})

	t.Run("malformed percent-encoding", func(t *testing.T) {
		if _, err := parseRedirectURL("/13/wall%"); err == nil {
			t.Error("expected error for malformed percent-encoding, got nil")
		}
	})
}

func TestAssembleTargetURL_VersionMapping(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantURL string
	}{
		{"version below 17 stays x.0.0", "/13/wall", "https://releases.llvm.org/13.0.0/tools/clang/docs/DiagnosticsReference.html#wall"},
		{"version 16 stays x.0.0", "/16/wall", "https://releases.llvm.org/16.0.0/tools/clang/docs/DiagnosticsReference.html#wall"},
		{"version 17 gets patchlevel 1", "/17/wall", "https://releases.llvm.org/17.0.1/tools/clang/docs/DiagnosticsReference.html#wall"},
		{"version 18 gets minorlevel 1", "/18/wall", "https://releases.llvm.org/18.1.0/tools/clang/docs/DiagnosticsReference.html#wall"},
		{"version 19 gets minorlevel 1", "/19/wall", "https://releases.llvm.org/19.1.0/tools/clang/docs/DiagnosticsReference.html#wall"},
		{"version 20 gets minorlevel 1", "/20/wall", "https://releases.llvm.org/20.1.0/tools/clang/docs/DiagnosticsReference.html#wall"},
		{"version 21 gets minorlevel 1", "/21/wall", "https://releases.llvm.org/21.1.0/tools/clang/docs/DiagnosticsReference.html#wall"},
		{"future version continues 18+ pattern", "/25/wall", "https://releases.llvm.org/25.1.0/tools/clang/docs/DiagnosticsReference.html#wall"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mustParseRedirectURL(t, tt.path)
			got, err := assembleTargetURL(u)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantURL {
				t.Errorf("got %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestAssembleTargetURL_FragmentNormalization(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantURL string
	}{
		{
			"plus signs become hyphens and double hyphens collapse",
			"/6/wc++98-c++11-compat-binary-literal",
			"https://releases.llvm.org/6.0.0/tools/clang/docs/DiagnosticsReference.html#wc-98-c-11-compat-binary-literal",
		},
		{
			"single plus sign",
			"/13/wc+98",
			"https://releases.llvm.org/13.0.0/tools/clang/docs/DiagnosticsReference.html#wc-98",
		},
		{
			"underscore passes through unchanged",
			"/13/W_underscore",
			"https://releases.llvm.org/13.0.0/tools/clang/docs/DiagnosticsReference.html#W_underscore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mustParseRedirectURL(t, tt.path)
			got, err := assembleTargetURL(u)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantURL {
				t.Errorf("got %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestAssembleTargetURL_Errors(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		errorContains string
	}{
		{"missing fragment", "/13", "insufficient parts"},
		{"empty path", "/", "insufficient parts"},
		{"non-numeric version", "/abc/wall", "version"},
		{"fragment with disallowed characters", "/13/wall!", "fragment"},
		{"version exceeding max length", "/" + strings.Repeat("1", 101) + "/wall", "version"},
		{"fragment exceeding max length", "/13/" + strings.Repeat("a", 51), "fragment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := parseRedirectURL(tt.path)
			if err != nil {
				t.Fatalf("parseRedirectURL(%q) returned unexpected error: %v", tt.path, err)
			}
			_, err = assembleTargetURL(u)
			if err == nil {
				t.Fatalf("expected error for path %q, got nil", tt.path)
			}
			if !strings.Contains(strings.ToLower(err.Error()), tt.errorContains) {
				t.Errorf("error %q does not contain expected substring %q", err.Error(), tt.errorContains)
			}
		})
	}
}

func TestAssembleTargetURL_ErrorMessagesAreEscaped(t *testing.T) {
	u := mustParseRedirectURL(t, "/<script>/wall")
	_, err := assembleTargetURL(u)
	if err == nil {
		t.Fatal("expected error for non-numeric version, got nil")
	}
	if strings.Contains(err.Error(), "<script>") {
		t.Errorf("error message contains unescaped input: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "&lt;script&gt;") {
		t.Errorf("error message does not contain escaped input: %q", err.Error())
	}
}

func TestAssembleTargetURL_EmptyInputPath(t *testing.T) {
	u := mustParseRedirectURL(t, "")
	if _, err := assembleTargetURL(u); err == nil {
		t.Error("expected error for empty input path, got nil")
	}
}
