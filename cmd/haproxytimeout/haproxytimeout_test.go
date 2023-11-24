package main_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	cmd "github.com/frobware/haproxytime/cmd/haproxytimeout"
)

// mockFailWriter is an io.Writer implementation that simulates a write failure.
// It's used in tests to trigger error handling paths in functions that write to io.Writer.
type mockFailWriter struct{}

func (m mockFailWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("mock write failure")
}

// MockExitHandler is an implementation of the ExitHandler interface used in tests.
// It captures the exit code provided to the Exit method instead of terminating the program.
type MockExitHandler struct {
	Exited bool // Exited indicates whether Exit was called
	Code   int  // Code is the exit code passed to Exit
}

func (e *MockExitHandler) Exit(code int) {
	e.Exited = true
	e.Code = code
}

type errorReader struct{}

func (er *errorReader) Read([]byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

type emptyStringReader struct {
	read bool
}

func (er *emptyStringReader) Read([]byte) (n int, err error) {
	if er.read {
		return 0, io.EOF
	}
	er.read = true
	return 0, nil
}

func TestVersion_Default(t *testing.T) {
	expected := ""
	if got := cmd.Version(); got != expected {
		t.Errorf("Version() = %v, want %v", got, expected)
	}
}

func TestVersion_Override(t *testing.T) {
	// Temporarily override Version
	originalVersion := cmd.Version
	defer func() { cmd.Version = originalVersion }()

	cmd.Version = func() string {
		return "test-version"
	}

	expected := "test-version"
	if got := cmd.Version(); got != expected {
		t.Errorf("Version() = %v, want %v", got, expected)
	}
}

func TestConvertDuration(t *testing.T) {
	tests := []struct {
		description    string
		args           []string
		stdin          io.Reader
		expectedExit   int
		expectedStdout string
		expectedStderr string
	}{{
		description:    "Test version flag",
		args:           []string{"-v"},
		expectedExit:   0,
		expectedStdout: "",
		expectedStderr: `haproxytimeout `,
	}, {
		description:    "Test -m flag",
		args:           []string{"-m"},
		expectedExit:   0,
		expectedStdout: "2147483647ms",
		expectedStderr: "",
	}, {
		description:    "Test -h flag",
		args:           []string{"-h", "2147483647ms"},
		expectedExit:   0,
		expectedStdout: "24d20h31m23s647ms",
		expectedStderr: "",
	}, {
		description:    "Test -m -h combined",
		args:           []string{"-m", "-h"},
		expectedExit:   0,
		expectedStdout: "24d20h31m23s647ms",
		expectedStderr: "",
	}, {
		description:    "number of milliseconds in a day from args",
		args:           []string{"86400000"},
		expectedExit:   0,
		expectedStdout: "86400000ms",
		expectedStderr: "",
	}, {
		description:    "number of milliseconds in a day from stdin",
		stdin:          strings.NewReader("1d\n"),
		expectedExit:   0,
		expectedStdout: "86400000ms",
		expectedStderr: "",
	}, {
		description:    "number of milliseconds in a day with human-readable output",
		args:           []string{"-h", "86400000"},
		expectedExit:   0,
		expectedStdout: "1d",
		expectedStderr: "",
	}, {
		description:    "1d as milliseconds",
		args:           []string{"1d"},
		expectedExit:   0,
		expectedStdout: "86400000ms",
		expectedStderr: "",
	}, {
		description:    "the HAProxy maximum duration",
		args:           []string{"-h", "24d20h31m23s646ms1000us"},
		expectedExit:   0,
		expectedStdout: "24d20h31m23s647ms",
		expectedStderr: "",
	}, {
		description:    "help flag",
		args:           []string{"-help"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: cmd.Usage,
	}, {
		description:    "single invalid flag",
		args:           []string{"-z"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "flag provided but not defined: -z",
	}, {
		description:    "mix of valid and invalid flags",
		args:           []string{"-h", "-z", "100ms"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "flag provided but not defined: -z",
	}, {
		description:    "syntax error reporting from args",
		args:           []string{"24d20h31m23s647msO000us"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "syntax error at position 18: invalid number\n24d20h31m23s647msO000us\n                 ^",
	}, {
		description:    "syntax error reporting from stdin",
		stdin:          strings.NewReader("24d20h31m23s647msO000us\n"),
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "syntax error at position 18: invalid number",
	}, {
		description:    "value exceeds HAProxy's maximum duration from args",
		args:           []string{"24d20h31m23s647ms1000us"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "range error at position 18\n24d20h31m23s647ms1000us\n                 ^",
	}, {
		description:    "value exceeds HAProxy's maximum description from stdin",
		stdin:          strings.NewReader("24d20h31m23s647ms1000us\n"),
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "range error at position 18",
	}, {
		description:    "simulated reading failure",
		stdin:          &errorReader{},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "error reading: simulated read error",
	}, {
		description:    "overflow error from args",
		args:           []string{"9223372036855ms"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "overflow error at position 1\n9223372036855ms\n^",
	}, {
		description:    "overflow error from stdin",
		stdin:          strings.NewReader("9223372036855ms"),
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "overflow error at position 1",
	}, {
		description:    "empty string from stdin",
		stdin:          &emptyStringReader{},
		expectedExit:   0,
		expectedStdout: "0ms",
		expectedStderr: "",
	}, {
		description:    "empty string from stdin with -h flag",
		args:           []string{"-h"},
		stdin:          &emptyStringReader{},
		expectedExit:   0,
		expectedStdout: "0ms",
		expectedStderr: "",
	}}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			mockExitHandler := &MockExitHandler{}

			exitCode := cmd.ConvertDuration(tc.stdin, stdout, stderr, tc.args, mockExitHandler)

			// If mockExitHandler.Exited is true, use
			// mockExitHandler.Code as the exit code.
			if mockExitHandler.Exited {
				exitCode = mockExitHandler.Code
			}

			if exitCode != tc.expectedExit {
				t.Errorf("Expected exit code %d, but got %d", tc.expectedExit, exitCode)
			}

			actualStdout := strings.TrimSuffix(stdout.String(), "\n")
			if actualStdout != tc.expectedStdout {
				t.Errorf("Expected stdout:\n<<<%s>>>\nBut got:\n<<<%s>>>", tc.expectedStdout, actualStdout)
			}

			actualStderr := strings.TrimSuffix(stderr.String(), "\n")
			if actualStderr != tc.expectedStderr {
				t.Errorf("Expected stderr:\n<<<%s>>>\nBut got:\n<<<%s>>>", tc.expectedStderr, actualStderr)
			}
		})
	}
}

// TestConvertDurationPrintFailure tests the convertDuration function
// to ensure it correctly handles write failures. The test uses
// mockFailWriter to simulate write failures and MockExitHandler to
// capture the exit behavior, verifying that convertDuration exits
// with the expected code when it encounters write errors.
func TestConvertDurationPrintFailure(t *testing.T) {
	mockStdin := strings.NewReader("1d")
	mockStdout := &mockFailWriter{}
	mockStderr := &mockFailWriter{}
	mockExitHandler := &MockExitHandler{}

	cmd.ConvertDuration(mockStdin, mockStdout, mockStderr, []string{}, mockExitHandler)

	// Verify that the mock exit handler was triggered with the
	// expected exit code.
	if !mockExitHandler.Exited || mockExitHandler.Code != 1 {
		t.Errorf("Expected exit with code 1, got exit %v with code %d", mockExitHandler.Exited, mockExitHandler.Code)
	}
}

// TestConvertDurationPrintlnFailure tests the convertDuration
// function to ensure it correctly handles write failures in
// safeFprintln. The test uses mockFailWriter to simulate write
// failures and MockExitHandler to capture the exit behavior. It
// verifies that convertDuration exits with the expected code when
// safeFprintln encounters write errors.
func TestConvertDurationPrintlnFailure(t *testing.T) {
	mockStdin := strings.NewReader("invalid input")
	mockStdout := &bytes.Buffer{}
	mockStderr := &mockFailWriter{}
	mockExitHandler := &MockExitHandler{}

	cmd.ConvertDuration(mockStdin, mockStdout, mockStderr, []string{"-h"}, mockExitHandler)

	// Verify that the mock exit handler was triggered with the
	// expected exit code.
	if !mockExitHandler.Exited || mockExitHandler.Code != 1 {
		t.Errorf("Expected exit with code 1, got exit %v with code %d", mockExitHandler.Exited, mockExitHandler.Code)
	}
}
