package main_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	cmd "github.com/frobware/haproxytime/cmd/haproxytimeout"
)

type errorReader struct{}

func (er *errorReader) Read([]byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

type emptyStringReader struct {
	read bool
}

func (er *emptyStringReader) Read(p []byte) (n int, err error) {
	if er.read {
		return 0, io.EOF
	}
	er.read = true
	return 0, nil
}

func TestConvertDuration(t *testing.T) {
	// We need a constant build version information for the tests
	// to pass.
	originalVersion := cmd.BuildVersion
	defer func() { cmd.BuildVersion = originalVersion }()
	cmd.BuildVersion = func() string {
		return "v0.0.0"
	}

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
		expectedStderr: `haproxytimeout: v0.0.0`,
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
		description:    "Almost the very very max",
		args:           []string{"-h", "24d20h31m23s646ms1000us"},
		expectedExit:   0,
		expectedStdout: "24d20h31m23s647ms",
		expectedStderr: "",
	}, {
		description:    "Test help flag",
		args:           []string{"-help"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: cmd.Usage,
	}, {
		description:    "Test single invalid flag",
		args:           []string{"-z"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "flag provided but not defined: -z",
	}, {
		description:    "Test mix of valid and invalid flags",
		args:           []string{"-h", "-z", "100ms"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "flag provided but not defined: -z",
	}, {
		description:    "Test syntax error reporting from args",
		args:           []string{"24d20h31m23s647msO000us"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "syntax error at position 18: invalid number\n24d20h31m23s647msO000us\n                 ^",
	}, {
		description:    "Test syntax error reporting from stdin",
		stdin:          strings.NewReader("24d20h31m23s647msO000us\n"),
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "syntax error at position 18: invalid number",
	}, {
		description:    "Test overflow error reporting from args",
		args:           []string{"24d20h31m23s647ms1000us"},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "overflow error at position 18: value exceeds max duration\n24d20h31m23s647ms1000us\n                 ^",
	}, {
		description:    "Test overflow error reporting from stdin",
		stdin:          strings.NewReader("24d20h31m23s647ms1000us\n"),
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "overflow error at position 18: value exceeds max duration",
	}, {
		description:    "Test simulated reading failure",
		stdin:          &errorReader{},
		expectedExit:   1,
		expectedStdout: "",
		expectedStderr: "error reading: simulated read error",
	}, {
		description:    "Test empty string from stdin",
		stdin:          &emptyStringReader{},
		expectedExit:   0,
		expectedStdout: "0ms",
		expectedStderr: "",
	}, {
		description:    "Test empty string from stdin with -h flag",
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

			exitCode := cmd.ConvertDuration(tc.stdin, stdout, stderr, tc.args)

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
