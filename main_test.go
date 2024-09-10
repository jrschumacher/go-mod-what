package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

const fixture = "fixtures/go.mod"
const nonexistentFixture = "fixtures/nonexistent_go.mod"
const invalidFixture = "fixtures/invalid_go.mod"

func TestMainFunction(t *testing.T) {
	// Save original os.Args and restore it after the test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Save original stdout and stderr and restore them after the test
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Create pipes to capture stdout and stderr output
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	tests := []struct {
		name      string
		args      []string
		modfile   string
		wantOut   string
		wantErr   string
		wantPanic bool
	}{
		{
			name:    "Help flag",
			args:    []string{"-help"},
			wantOut: usage,
		},
		{
			name:    "No modfile provided",
			args:    []string{"-modfile", ""},
			wantErr: "go.mod file not provided\n\n" + usage + "\n" + usageExample + "\n",
		},
		{
			name:    "Invalid modfile path",
			args:    []string{"-modfile", nonexistentFixture},
			wantErr: "failed to read go.mod file: open invalid/path/go.mod: no such file or directory\n\n",
		},
		{
			name:    "Invalid modfile content",
			args:    []string{"-modfile", invalidFixture},
			wantErr: "failed to parse go.mod file: testdata/invalid_go.mod:3: invalid module path: invalid content\n\n",
		},
		{
			name:    "Package not found",
			args:    []string{"github.com/nonexistent/package"},
			wantErr: "github.com/nonexistent/package not found\n",
		},
		{
			name:    "Package found",
			args:    []string{"-modfile", fixture, "golang.org/x/a"},
			wantOut: "golang.org/x/a	v0.0.1\n",
		},
		{
			name:    "Package found with only-version",
			args:    []string{"-modfile", fixture, "-only-version", "golang.org/x/a"},
			wantOut: "v0.0.1\n",
		},
	}

	for _, tt := range tests {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		t.Run(tt.name, func(t *testing.T) {
			// Set os.Args for the test
			os.Args = append([]string{"cmd"}, tt.args...)

			// Write modfile content if provided
			if tt.modfile != "" {
				os.WriteFile("testdata/go.mod", []byte(tt.modfile), 0644)
				defer os.Remove("testdata/go.mod")
			}

			// Run main function and capture panic if any
			defer func() {
				if r := recover(); r != nil {
					if tt.wantErr == "" {
						t.Errorf("main() panicked: %v", r)
					}
				}
			}()

			main()

			// Close writers and read captured output
			wOut.Close()
			wErr.Close()
			var bufOut, bufErr bytes.Buffer
			bufOut.ReadFrom(rOut)
			bufErr.ReadFrom(rErr)

			// Check stdout
			if gotOut := bufOut.String(); !strings.Contains(gotOut, tt.wantOut) {
				if tt.wantOut == "" && gotOut == "" {
					return
				}
				t.Errorf("stdout = %v, want %v", gotOut, tt.wantOut)
			}

			// Check stderr
			if gotErr := bufErr.String(); !strings.Contains(gotErr, tt.wantErr) {
				if tt.wantErr == "" && gotErr == "" {
					return
				}
				t.Errorf("stderr = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	t.Setenv("TEST_MODE", "true")

	// Save original stderr and restore it after the test
	origStderr := os.Stderr
	defer func() { os.Stderr = origStderr }()

	// Create a pipe to capture stderr output
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test cases
	tests := []struct {
		name    string
		message string
		err     error
		with    []bool
		want    string
	}{
		{
			name:    "Error with message only",
			message: "test error",
			err:     nil,
			with:    nil,
			want:    "test error\n\n",
		},
		{
			name:    "Error with message and error",
			message: "test error",
			err:     fmt.Errorf("an error occurred"),
			with:    nil,
			want:    "test error: an error occurred\n\n",
		},
		{
			name:    "Error with message, error, and usage",
			message: "test error",
			err:     fmt.Errorf("an error occurred"),
			with:    []bool{true},
			want:    "test error: an error occurred\n\n" + usage + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// recover from panic to simulate os.Exit
			defer func() {
				if r := recover(); r != nil {
					fmt.Print("Recovered from panic: ", r)
				}
			}()

			// Reset flag.Usage to avoid printing usage multiple times
			flag.Usage = func() {}

			// Call printError
			printError(tt.message, tt.err, tt.with...)

			// Close the writer and read the captured output
			w.Close()
			var buf bytes.Buffer
			buf.ReadFrom(r)

			// Check if the output matches the expected output
			if got := buf.String(); got != tt.want {
				t.Errorf("printError() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestCompareRequire(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "Exact match",
			a:    "github.com/gorilla/mux",
			b:    "github.com/gorilla/mux",
			want: true,
		},
		{
			name: "Wildcard match",
			a:    "github.com/gorilla/*",
			b:    "github.com/gorilla/mux",
			want: true,
		},
		{
			name: "No match",
			a:    "github.com/gorilla/schema",
			b:    "github.com/gorilla/mux",
			want: false,
		},
		{
			name: "Wildcard no match",
			a:    "github.com/gorilla/*",
			b:    "github.com/another/package",
			want: false,
		},
		{
			name: "Partial match",
			a:    "github.com/gorilla",
			b:    "github.com/gorilla/mux",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareRequire(tt.a, tt.b); got != tt.want {
				t.Errorf("compareRequire(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
