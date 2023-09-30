package haproxytime_test

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/frobware/haproxytime"
)

// TestSyntaxError_Error ensures that the error message produced by a
// SyntaxError provides accurate and clear information about the
// syntax issue encountered. The test crafts a duration string known
// to cause a syntax error and then verifies that the SyntaxError
// generated details the correct position and nature of the error.
func TestSyntaxError_Error(t *testing.T) {
	tests := []struct {
		input            string
		expectedPosition int
		expectedCause    haproxytime.SyntaxErrorCause
		expectedErrorMsg string
		parseMode        haproxytime.ParseMode
	}{{
		input:            "1h1x",
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnit,
		expectedErrorMsg: "syntax error at position 4: invalid unit",
		parseMode:        haproxytime.ParseModeMultiUnit,
	}, {
		input:            "xx1h",
		expectedPosition: 1,
		expectedCause:    haproxytime.InvalidNumber,
		expectedErrorMsg: "syntax error at position 1: invalid number",
		parseMode:        haproxytime.ParseModeMultiUnit,
	}, {
		input:            "1m1h",
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		expectedErrorMsg: "syntax error at position 4: invalid unit order",
		parseMode:        haproxytime.ParseModeMultiUnit,
	}, {
		input:            "1h1m1h",
		expectedPosition: 3,
		expectedCause:    haproxytime.UnexpectedCharactersInSingleUnitMode,
		expectedErrorMsg: "syntax error at position 3: unexpected characters in single unit mode",
		parseMode:        haproxytime.ParseModeSingleUnit,
	}}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, err := haproxytime.ParseDuration(tc.input, haproxytime.UnitMillisecond, tc.parseMode)

			if !errors.Is(err, &haproxytime.SyntaxError{}) {
				t.Errorf("Expected a SyntaxError, but got %T", err)
				return
			}

			syntaxErr := err.(*haproxytime.SyntaxError)

			if syntaxErr.Error() != tc.expectedErrorMsg {
				t.Errorf("expected %q, but got %q", tc.expectedErrorMsg, syntaxErr.Error())
			}

			if syntaxErr.Position() != tc.expectedPosition-1 {
				t.Errorf("expected SyntaxError at position %v, but got %v", tc.expectedPosition-1, syntaxErr.Position())
			}

			if syntaxErr.Cause() != tc.expectedCause {
				t.Errorf("expected SyntaxError cause to be %v, but got %v", tc.expectedCause, syntaxErr.Cause())
			}
		})
	}
}

// TestOverflowError_Error validates that the error message produced
// by an OverflowError accurately represents the cause of the
// overflow. The test crafts a duration string known to cause an
// overflow, then ensures that the OverflowError generated reports the
// correct position and value causing the overflow.
func TestOverflowError_Error(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    string
	}{{
		description: "overflows haproxy max duration",
		input:       "2147483648ms",
		expected:    "overflow error at position 1: value exceeds max duration",
	}, {
		description: "100 days overflows haproxy max duration",
		input:       "100d",
		expected:    "overflow error at position 1: value exceeds max duration",
	}, {
		description: "genuine int64 range error",
		input:       "10000000000000000000",
		expected:    "overflow error at position 1: value exceeds max duration",
	}}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			_, err := haproxytime.ParseDuration(tc.input, haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit)
			if !errors.Is(err, &haproxytime.OverflowError{}) {
				t.Errorf("expected OverflowError, got %T", err)
				return
			}

			overflowErr := err.(*haproxytime.OverflowError)

			if overflowErr.Error() != tc.expected {
				t.Errorf("expected %q, but got %q", tc.expected, overflowErr.Error())
			}
		})
	}
}

// TestParseDurationOverflowErrors verifies the proper handling of
// overflow errors when parsing duration strings. It ensures that
// values within the acceptable range do not produce errors, while
// values exceeding the limits are correctly identified and reported
// with detailed information, including the problematic number and its
// position within the input.
func TestParseDurationOverflowErrors(t *testing.T) {
	tests := []struct {
		description    string
		input          string
		expectErr      bool
		expectedErrPos int
		duration       time.Duration
	}{{
		description: "2147483648us is within limits",
		input:       "2147483648us",
		expectErr:   false,
		duration:    2147483648 * time.Microsecond,
	}, {
		description: "maximum value without overflow (just under the limit)",
		input:       "2147483647ms",
		expectErr:   false,
		duration:    haproxytime.MaxTimeout,
	}, {
		description: "maximum value without overflow (using different time units)",
		input:       "2147483s647ms",
		expectErr:   false,
		duration:    2147483*time.Second + 647*time.Millisecond,
	}, {
		description: "maximum value without overflow (using different time units)",
		input:       "35791m23s647ms",
		expectErr:   false,
		duration:    35791*time.Minute + 23*time.Second + 647*time.Millisecond,
	}, {
		description: "way below the limit",
		input:       "1000ms",
		expectErr:   false,
		duration:    1000 * time.Millisecond,
	}, {
		description: "way below the limit",
		input:       "1s",
		expectErr:   false,
		duration:    1 * time.Second,
	}, {
		description: "MaxInt32 milliseconds",
		input:       fmt.Sprintf("%dms", math.MaxInt32),
		expectErr:   false,
		duration:    time.Duration(math.MaxInt32) * time.Millisecond,
	}, {
		description:    "just over the limit",
		input:          "2147483648ms",
		expectErr:      true,
		expectedErrPos: 1,
		duration:       0,
	}, {
		description:    "over the limit with combined units",
		input:          "2147483s648ms",
		expectErr:      true,
		expectedErrPos: 9,
		duration:       0,
	}, {
		description:    "over the limit (using different time units)",
		input:          "35791m23s648ms",
		expectErr:      true,
		expectedErrPos: 10,
		duration:       0,
	}, {
		description:    "way over the limit",
		input:          "4294967295ms",
		expectErr:      true,
		expectedErrPos: 1,
		duration:       0,
	}, {
		description:    "way, way over the limit",
		input:          "9223372036854775807d",
		duration:       0,
		expectErr:      true,
		expectedErrPos: 1,
	}, {
		description:    "maximum value +1 (using different units)",
		input:          "24d20h31m23s648ms0us",
		expectErr:      true,
		expectedErrPos: 13,
		duration:       0,
	}, {
		description:    "MaxInt32+1 milliseconds",
		input:          fmt.Sprintf("%vms", math.MaxInt32+1),
		expectErr:      true,
		expectedErrPos: 1,
		duration:       0,
	}, {
		description:    "MaxInt64 milliseconds",
		input:          fmt.Sprintf("%vms", math.MaxInt64),
		expectErr:      true,
		expectedErrPos: 1,
		duration:       0,
	}}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			duration, err := haproxytime.ParseDuration(tc.input, haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit)

			if tc.expectErr {
				if !errors.Is(err, &haproxytime.OverflowError{}) {
					t.Errorf("Expected an OverflowError, but got %T", err)
					return
				}

				overflowErr := err.(*haproxytime.OverflowError)

				// Use a 1-based index in the test
				// case to avoid inadvertently relying
				// on the zero value for
				// expectedPosition.
				if overflowErr.Position() != tc.expectedErrPos-1 {
					t.Errorf("Expected OverflowError at position %v, but got %v", tc.expectedErrPos-1, overflowErr.Position())
				}
			} else {
				if err != nil {
					t.Errorf("Didn't expect error for input %q but got %v", tc.input, err)
					return
				}
				if duration != tc.duration {
					t.Errorf("expected duration %v, but got %v", tc.duration, duration)
				}
			}
		})
	}
}

// TestParseDurationSyntaxErrors verifies that duration strings are
// parsed correctly according to their syntax. It checks various valid
// and invalid inputs to ensure the parser handles syntax errors
// appropriately, identifying and reporting any inconsistencies or
// unsupported formats with a detailed error message and the position
// of the problematic segment.
func TestParseDurationSyntaxErrors(t *testing.T) {
	tests := []struct {
		description      string
		input            string
		expectErr        bool
		expectedPosition int
		expectedCause    haproxytime.SyntaxErrorCause
		duration         time.Duration
	}{{
		description: "empty string",
		input:       "",
		expectErr:   false,
		duration:    0,
	}, {
		description: "zero milliseconds",
		input:       "0",
		expectErr:   false,
		duration:    0,
	}, {
		description: "all units specified",
		input:       "1d3h30m45s100ms200us",
		expectErr:   false,
		duration:    27*time.Hour + 30*time.Minute + 45*time.Second + 100*time.Millisecond + 200*time.Microsecond,
	}, {
		description: "default unit",
		input:       "5000",
		expectErr:   false,
		duration:    5000 * time.Millisecond,
	}, {
		description: "number with leading zeros",
		input:       "0101us",
		expectErr:   false,
		duration:    101 * time.Microsecond,
	}, {
		description: "zero milliseconds",
		input:       "0ms",
		expectErr:   false,
		duration:    0,
	}, {
		description: "all units as zero",
		input:       "0d0h0m0s0ms0us",
		expectErr:   false,
		duration:    0,
	}, {
		description: "all units as zero with implicit milliseconds",
		input:       "0d0h0m0s00us",
		expectErr:   false,
		duration:    0,
	}, {
		description: "1 millisecond",
		input:       "0d0h0m0s1",
		expectErr:   false,
		duration:    time.Millisecond,
	}, {
		description: "skipped units",
		input:       "1d100us",
		expectErr:   false,
		duration:    24*time.Hour + 100*time.Microsecond,
	}, {
		description: "maximum number of ms",
		input:       "2147483647",
		expectErr:   false,
		duration:    2147483647 * time.Millisecond,
	}, {
		description: "maximum number expressed with all units",
		input:       "24d20h31m23s647ms0us",
		expectErr:   false,
		duration:    2147483647 * time.Millisecond,
	}, {
		description:      "leading space is not a number",
		input:            " ",
		expectErr:        true,
		expectedPosition: 1,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "leading +",
		input:            "+0",
		expectErr:        true,
		expectedPosition: 1,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "negative number",
		input:            "-1",
		expectErr:        true,
		expectedPosition: 1,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "abc is an invalid number",
		input:            "abc",
		expectErr:        true,
		expectedPosition: 1,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "/ is an invalid number",
		input:            "/",
		expectErr:        true,
		expectedPosition: 1,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      ". is an invalid unit",
		input:            "100.d",
		expectErr:        true,
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnit,
		duration:         0,
	}, {
		description:      "X is an invalid number after the valid 1d30m",
		input:            "1d30mX",
		expectErr:        true,
		expectedPosition: 6,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "Y is an invalid unit after the valid 2d30m and the next digit",
		input:            "2d30m1Y",
		expectErr:        true,
		expectedPosition: 7,
		expectedCause:    haproxytime.InvalidUnit,
		duration:         0,
	}, {
		description:      "duplicate day unit",
		input:            "1d1d",
		expectErr:        true,
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate hour unit",
		input:            "2d1h1h",
		expectErr:        true,
		expectedPosition: 6,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate minute unit",
		input:            "3d2h1m1m",
		expectErr:        true,
		expectedPosition: 8,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate second unit",
		input:            "4d3h2m1s1s",
		expectErr:        true,
		expectedPosition: 10,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate millisecond unit",
		input:            "5d4h3m2s1ms1ms",
		expectErr:        true,
		expectedPosition: 13,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate microsecond unit",
		input:            "6d5h4m3s2ms1us1us",
		expectErr:        true,
		expectedPosition: 16,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "milliseconds cannot follow microseconds",
		input:            "5us1ms",
		expectErr:        true,
		expectedPosition: 5,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "seconds cannot follow milliseconds",
		input:            "5ms1s",
		expectErr:        true,
		expectedPosition: 5,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "minutes cannot follow seconds",
		input:            "5s1m",
		expectErr:        true,
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "hours cannot cannot follow minutes",
		input:            "5m1h",
		expectErr:        true,
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "days cannot folow hours",
		input:            "5h1d",
		expectErr:        true,
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			duration, err := haproxytime.ParseDuration(tc.input, haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit)

			if tc.expectErr {
				if !errors.Is(err, &haproxytime.SyntaxError{}) {
					t.Errorf("Expected a SyntaxError, but got %T", err)
					return
				}

				syntaxErr := err.(*haproxytime.SyntaxError)

				// Use a 1-based index in the test
				// case to avoid inadvertently relying
				// on the zero value for
				// expectedPosition.
				if syntaxErr.Position() != tc.expectedPosition-1 {
					t.Errorf("Expected position %v, but got %v", tc.expectedPosition-1, syntaxErr.Position())
				}
				if syntaxErr.Cause() != tc.expectedCause {
					t.Errorf("Expected cause %v, but got %v", tc.expectedCause, syntaxErr.Cause())
				}
			} else {
				if err != nil {
					t.Errorf("Didn't expect error for input %q but got %v", tc.input, err)
					return
				}
				if duration != tc.duration {
					t.Errorf("expected duration %v, but got %v", tc.duration, duration)
				}
			}
		})
	}
}
