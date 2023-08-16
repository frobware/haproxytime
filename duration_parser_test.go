package haproxytime_test

import (
	"errors"
	"fmt"
	"math"
	"os"
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
		input:          "9223372036855ms",
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
				// case to avoid relying on the zero
				// value. Adjust by subtracting 1 when
				// comparing positions.
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
		description:      "duplicate units",
		input:            "0ms0ms",
		expectErr:        true,
		expectedPosition: 5,
		expectedCause:    haproxytime.InvalidUnitOrder,
		duration:         0,
	}, {
		description:      "out of order units, hours cannot follow minutes",
		input:            "1d5m1h",
		expectErr:        true,
		expectedPosition: 6,
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
				// case to avoid relying on the zero
				// value. Adjust by subtracting 1 when
				// comparing positions.
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

// How to interpret the benchmark results for:
//
// $ go test -bench=. -benchmem [-count=1] [-benchtime=1s]
// goos: linux
// goarch: amd64
// pkg: github.com/frobware/haproxytime
// cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
// BenchmarkParseDurationMultiUnitMode-8   7338422 168.1 ns/op 0 B/op 0 allocs/op
// BenchmarkParseDurationSingleUnitMode-8 28579370 41.55 ns/op 0 B/op 0 allocs/op
// PASS
// ok github.com/frobware/haproxytime 2.648s
//
// Here's a breakdown:
//
//   - `BenchmarkParseDurationMultiUnitMode-8`: This tells you the
//     name of the benchmark function that was executed. The `-8`
//     specifies that the benchmark was run with 8 threads.
//
//   - `7338422`: This is the number of iterations that the benchmark
//     managed to run during its timed execution.
//
//   - `168.1 ns/op`: This tells you that each operation (in this case,
//     a call to `ParseDuration`) took an average of 168.1 nanoseconds.
//
//   - `0 B/op`: This indicates that the function did not allocate any
//     additional bytes of memory per operation. This is often a
//     crucial factor in performance-sensitive code, so having 0 here
//     is generally a good sign.
//
//   - `0 allocs/op`: This tells you that the function did not make
//     any heap allocations per operation. Fewer allocations often lead
//     to faster code and less pressure on the garbage collector, so
//     this is also a positive indicator.
//
//   - `PASS`: This tells you that the benchmark completed successfully
//     without any errors.
//
//   - `ok github.com/frobware/haproxytime 2.648s`: This indicates that
//     the entire test, including setup, tear-down, and the running of
//     the benchmark, completed in 2.648s.

func BenchmarkParseDurationMultiUnitMode(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := haproxytime.ParseDuration("24d20h31m23s647ms", haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDurationSingleUnitMode(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := haproxytime.ParseDuration("2147483647ms", haproxytime.UnitMillisecond, haproxytime.ParseModeSingleUnit)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// This example gets copied to the documentation.
func ExampleParseDuration() {
	if duration, err := haproxytime.ParseDuration("24d20h31m23s647ms", haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("%vms\n", duration.Milliseconds())
	}
	// Output: 2147483647ms
}
