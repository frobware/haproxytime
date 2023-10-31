package haproxytime_test

import (
	"errors"
	"testing"
	"time"

	"github.com/frobware/haproxytime"
)

// TestSynxtaxError_Error validates that the error message produced by
// a SyntaxError accurately represents the position, which is 1-based,
// and the cause of the error.
func TestSyntaxError_Error(t *testing.T) {
	tests := []struct {
		input            string
		expectedErrorMsg string
		expectedPosition int
		expectedCause    haproxytime.SyntaxErrorCause
		parseMode        haproxytime.ParseMode
	}{{
		input:            "1h1x",
		expectedErrorMsg: "syntax error at position 4: invalid unit",
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnit,
		parseMode:        haproxytime.ParseModeMultiUnit,
	}, {
		input:            "xx1h",
		expectedErrorMsg: "syntax error at position 1: invalid number",
		expectedPosition: 1,
		expectedCause:    haproxytime.InvalidNumber,
		parseMode:        haproxytime.ParseModeMultiUnit,
	}, {
		input:            "1m1h",
		expectedErrorMsg: "syntax error at position 4: invalid unit order",
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		parseMode:        haproxytime.ParseModeMultiUnit,
	}, {
		input:            "1h1m1h",
		expectedErrorMsg: "syntax error at position 3: unexpected characters in single unit mode",
		expectedPosition: 3,
		expectedCause:    haproxytime.UnexpectedCharactersInSingleUnitMode,
		parseMode:        haproxytime.ParseModeSingleUnit,
	}}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, err := haproxytime.ParseDuration(tc.input, haproxytime.Millisecond, tc.parseMode, haproxytime.NoRangeChecking)

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
// overflow.
func TestOverflowError_Error(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{{
		input:    "106751d23h47m16s854ms984us",
		expected: "overflow error at position 22",
	}, {
		input:    "9223372036854775808ns",
		expected: "overflow error at position 1",
	}}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			_, err := haproxytime.ParseDuration(tc.input, haproxytime.Microsecond, haproxytime.ParseModeMultiUnit, haproxytime.NoRangeChecking)
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

// TestParseDurationOverflow validates the parsing logic across a
// range of input strings representing various time durations. The
// function takes into account boundary cases and the maximum
// representable duration.
//
// The maximum duration that can be represented in a time.Duration
// value is determined by the limits of int64, as time.Duration is
// just an alias for int64 where each unit represents a nanosecond.
//
// The maximum int64 value is 9223372036854775807.
//
// Therefore, the maximum durations for various units are calculated
// as follows:
//
// - Nanoseconds: 9223372036854775807 (since the base unit is a nanosecond)
// - Microseconds: 9223372036854775 (9223372036854775807 / 1000)
// - Milliseconds: 9223372036854 (9223372036854775807 / 1000000)
// - Seconds: 9223372036 (9223372036854775807 / 1000000000)
// - Minutes: 153722867 (9223372036854775807 / 60000000000)
// - Hours: 2562047 (9223372036854775807 / 3600000000000)
// - Days: 106751 (9223372036854775807 / 86400000000000)
func TestParseDurationOverflow(t *testing.T) {
	tests := []struct {
		input            string
		expectedDuration time.Duration
		expectedErrPos   int
		expectErr        bool
	}{{
		input:            "",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "0",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "0us",
		expectedDuration: 0 * time.Microsecond,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "1us",
		expectedDuration: 1 * time.Microsecond,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "0ms",
		expectedDuration: 0 * time.Millisecond,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "1ms",
		expectedDuration: 1 * time.Millisecond,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "0s",
		expectedDuration: 0 * time.Second,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "1s",
		expectedDuration: 1 * time.Second,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "0m",
		expectedDuration: 0 * time.Minute,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "1m",
		expectedDuration: 1 * time.Minute,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "0h",
		expectedDuration: 0 * time.Hour,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "1h",
		expectedDuration: 1 * time.Hour,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "0d",
		expectedDuration: 0 * time.Hour,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "1d",
		expectedDuration: 24 * time.Hour,
		expectedErrPos:   0,
		expectErr:        false,
	}, { // The largest representable value using a single unit.
		input:            "9223372036854775us",
		expectedDuration: 9223372036854775 * time.Microsecond,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "9223372036854ms",
		expectedDuration: 9223372036854 * time.Millisecond,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "9223372036s",
		expectedDuration: 9223372036 * time.Second,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "153722867m",
		expectedDuration: 153722867 * time.Minute,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "2562047h",
		expectedDuration: 2562047 * time.Hour,
		expectedErrPos:   0,
		expectErr:        false,
	}, {
		input:            "106751d",
		expectedDuration: 106751 * 24 * time.Hour,
		expectedErrPos:   0,
		expectErr:        false,
	}, { // The largest representable value using mixed units.
		input:            "106751d5h59m59s854ms775us",
		expectedDuration: 106751*24*time.Hour + 5*time.Hour + 59*time.Minute + 59*time.Second + 854*time.Millisecond + 775*time.Microsecond,
		expectedErrPos:   0,
		expectErr:        false,
	}, { // Overflow cases.
		input:            "9223372036854776us",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        true,
	}, {
		input:            "9223372036855ms",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        true,
	}, {
		input:            "9223372037s",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        true,
	}, {
		input:            "153722868m",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        true,
	}, {
		input:            "2562048h",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        true,
	}, {
		input:            "106752d",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        true,
	}, { // Overflow with all units.
		input:            "106751d23h47m16s854ms984us",
		expectedDuration: 0,
		expectedErrPos:   21,
		expectErr:        true,
	}, { // max Int64 +1
		input:            "9223372036854775808",
		expectedDuration: 0,
		expectedErrPos:   0,
		expectErr:        true,
	}}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			duration, err := haproxytime.ParseDuration(tc.input, haproxytime.Microsecond, haproxytime.ParseModeMultiUnit, haproxytime.NoRangeChecking)

			if tc.expectErr {
				if !errors.Is(err, &haproxytime.OverflowError{}) {
					t.Errorf("expected an OverflowError, but got %T", err)
					return
				}

				overflowErr := err.(*haproxytime.OverflowError)
				if overflowErr.Position() != tc.expectedErrPos {
					t.Errorf("expected OverflowError at position %v, but got %v", tc.expectedErrPos, overflowErr.Position())
				}
			} else {
				if err != nil {
					t.Errorf("didn't expect an error for input %q but got %v", tc.input, err)
					return
				}
				if duration != tc.expectedDuration {
					t.Errorf("expected duration %v, but got %v", tc.expectedDuration, duration)
				}
			}
		})
	}
}

// TestParseDurationSyntaxErrors verifies that duration strings are
// parsed correctly according to their syntax. It checks various valid
// and invalid inputs to ensure the parser handles syntax errors
// appropriately, identifying and reporting any inconsistencies or
// unsupported formats.
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
		expectedPosition: 0,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "leading +",
		input:            "+0",
		expectErr:        true,
		expectedPosition: 0,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "negative number",
		input:            "-1",
		expectErr:        true,
		expectedPosition: 0,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "abc is an invalid number",
		input:            "abc",
		expectErr:        true,
		expectedPosition: 0,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "/ is an invalid number",
		input:            "/",
		expectErr:        true,
		expectedPosition: 0,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      ". is an invalid unit",
		input:            "100.d",
		expectErr:        true,
		expectedPosition: 3,
		expectedCause:    haproxytime.InvalidUnit,
		duration:         0,
	}, {
		description:      "X is an invalid number after the valid 1d30m",
		input:            "1d30mX",
		expectErr:        true,
		expectedPosition: 5,
		expectedCause:    haproxytime.InvalidNumber,
		duration:         0,
	}, {
		description:      "Y is an invalid unit after the valid 2d30m and the next digit",
		input:            "2d30m1Y",
		expectErr:        true,
		expectedPosition: 6,
		expectedCause:    haproxytime.InvalidUnit,
		duration:         0,
	}, {
		description:      "duplicate day unit",
		input:            "1d1d",
		expectErr:        true,
		expectedPosition: 3,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate hour unit",
		input:            "2d1h1h",
		expectErr:        true,
		expectedPosition: 5,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate minute unit",
		input:            "3d2h1m1m",
		expectErr:        true,
		expectedPosition: 7,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate second unit",
		input:            "4d3h2m1s1s",
		expectErr:        true,
		expectedPosition: 9,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate millisecond unit",
		input:            "5d4h3m2s1ms1ms",
		expectErr:        true,
		expectedPosition: 12,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "duplicate microsecond unit",
		input:            "6d5h4m3s2ms1us1us",
		expectErr:        true,
		expectedPosition: 15,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "milliseconds cannot follow microseconds",
		input:            "5us1ms",
		expectErr:        true,
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "seconds cannot follow milliseconds",
		input:            "5ms1s",
		expectErr:        true,
		expectedPosition: 4,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "minutes cannot follow seconds",
		input:            "5s1m",
		expectErr:        true,
		expectedPosition: 3,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "hours cannot cannot follow minutes",
		input:            "5m1h",
		expectErr:        true,
		expectedPosition: 3,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "days cannot folow hours",
		input:            "5h1d",
		expectErr:        true,
		expectedPosition: 3,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			duration, err := haproxytime.ParseDuration(tc.input, haproxytime.Millisecond, haproxytime.ParseModeMultiUnit, haproxytime.NoRangeChecking)

			if tc.expectErr {
				if !errors.Is(err, &haproxytime.SyntaxError{}) {
					t.Errorf("expected a SyntaxError, but got %T", err)
					return
				}

				syntaxErr := err.(*haproxytime.SyntaxError)
				if syntaxErr.Position() != tc.expectedPosition {
					t.Errorf("expected SyntaxError at position %v, but got %v", tc.expectedPosition, syntaxErr.Position())
				}
				if syntaxErr.Cause() != tc.expectedCause {
					t.Errorf("expected cause %v, but got %v", tc.expectedCause, syntaxErr.Cause())
				}
			} else {
				if err != nil {
					t.Errorf("didn't expect an error for input %q but got %v", tc.input, err)
					return
				}
				if duration != tc.duration {
					t.Errorf("expected duration %v, but got %v", tc.duration, duration)
				}
			}
		})
	}
}

func TestParseDurationRangeErrors(t *testing.T) {
	tests := []struct {
		description      string
		input            string
		expectErr        bool
		expectedPosition int
		expectedErrMsg   string
		duration         time.Duration
		inRangeChecker   haproxytime.RangeChecker
	}{{
		description:    "empty string",
		input:          "",
		expectErr:      false,
		expectedErrMsg: "",
		duration:       0,
		inRangeChecker: func(position int, value time.Duration, total time.Duration) bool {
			return false
		},
	}, {
		description:      "0 is out of range",
		input:            "0",
		expectErr:        true,
		expectedPosition: 0,
		expectedErrMsg:   "range error at position 1",
		duration:         0,
		inRangeChecker: func(position int, value time.Duration, total time.Duration) bool {
			return false
		},
	}, {
		description:      "the final 1000us takes the input out of range",
		input:            "24d20h31m23s647ms1000us",
		expectErr:        true,
		expectedPosition: 17,
		expectedErrMsg:   "range error at position 18",
		duration:         0,
		inRangeChecker: func(position int, value time.Duration, total time.Duration) bool {
			return total.Milliseconds() < 2147483647
		},
	}}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			duration, err := haproxytime.ParseDuration(tc.input, haproxytime.Millisecond, haproxytime.ParseModeMultiUnit, tc.inRangeChecker)

			if tc.expectErr {
				if !errors.Is(err, &haproxytime.RangeError{}) {
					t.Errorf("expected a RangeError, but got %T", err)
					return
				}

				rangeErr := err.(*haproxytime.RangeError)
				if rangeErr.Position() != tc.expectedPosition {
					t.Errorf("expected RangeError at position %v, but got %v", tc.expectedPosition, rangeErr.Position())
				}
				if rangeErr.Error() != tc.expectedErrMsg {
					t.Errorf("expected error message %q, got %q", tc.expectedErrMsg, rangeErr.Error())
				}
			} else {
				if err != nil {
					t.Errorf("didn't expect an error for input %q but got %v", tc.input, err)
					return
				}
				if duration != tc.duration {
					t.Errorf("expected duration %v, but got %v", tc.duration, duration)
				}
			}
		})
	}
}
