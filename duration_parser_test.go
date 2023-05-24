package haproxytime_test

import (
	"testing"
	"time"

	"github.com/frobware/haproxytime"
)

func TestParseDuration(t *testing.T) {
	testCases := []struct {
		description string
		input       string
		duration    time.Duration
		error       string
	}{
		{
			description: "test with empty string",
			input:       "",
			duration:    0,
		},
		{
			description: "test with string that is just spaces",
			input:       "   ",
			duration:    0,
		},
		{
			description: "test for zero",
			input:       "0",
			duration:    0,
		},
		{
			description: "invalid number",
			input:       "a",
			error:       "invalid number",
		},
		{
			description: "invalid number",
			input:       "/",
			error:       "invalid number",
		},
		{
			description: "invalid number, because the 100 defaults to 100ms",
			input:       "100.d",
			error:       "invalid number",
		},
		{
			description: "invalid unit",
			input:       "1d 30mgarbage",
			error:       "invalid unit",
		},
		{
			description: "valid test with spaces",
			input:       "1d 3h 30m 45s 100ms 200us",
			duration:    time.Hour*27 + time.Minute*30 + time.Second*45 + time.Millisecond*100 + time.Microsecond*200,
		},
		{
			description: "valid test with no space",
			input:       "1d3h30m45s100ms200us",
			duration:    27*time.Hour + 30*time.Minute + 45*time.Second + 100*time.Millisecond + 200*time.Microsecond,
		},
		{
			description: "test with leading and trailing spaces",
			input:       "  1d   3h   30m   45s   ",
			duration:    27*time.Hour + 30*time.Minute + 45*time.Second,
		},
		{
			description: "test with no unit (assume milliseconds)",
			input:       "5000",
			duration:    5000 * time.Millisecond,
		},
		{
			description: "test with no unit (assume milliseconds), and repeated milliseconds",
			input:       "5000 100ms",
			error:       "invalid unit order",
		},
		{
			description: "test with no unit (assume milliseconds), followed by another millisecond value",
			input:       "5000 100ms",
			error:       "invalid unit order",
		},
		{
			description: "test number with leading zeros",
			input:       "000000000000000000000001 01us",
			duration:    time.Millisecond + time.Microsecond,
		},
		{
			description: "test for zero milliseconds",
			input:       "0ms",
			duration:    0,
		},
		{
			description: "test all units as zero",
			input:       "0d 0h 0m 0s 0ms 0us",
			duration:    0,
		},
		{
			description: "test all units as zero with implicit milliseconds",
			input:       "0d 0h 0m 0s 0 0us",
			duration:    0,
		},
		{
			description: "test with all zeros, and trailing 0 with no unit but ms has already been specified",
			input:       "0d 0h 0m 0s 0ms 0",
			error:       "invalid unit order",
		},
		{
			description: "test 1 millisecond",
			input:       "0d 0h 0m 0s 1",
			duration:    time.Millisecond,
		},
		{
			description: "test duplicate units",
			input:       "0ms 0ms",
			error:       "invalid unit order",
		},
		{
			description: "test out of order units, hours cannot follow minutes",
			input:       "1d 5m 1h",
			error:       "invalid unit order",
		},
		{
			description: "test skipped units",
			input:       "1d 100us",
			duration:    24*time.Hour + 100*time.Microsecond,
		},
		{
			description: "test maximum number of seconds",
			input:       "9223372036s",
			duration:    9223372036 * time.Second,
		},
		{
			description: "test overflow",
			input:       "9223372036s 1000ms",
			error:       "overflow",
		},
		{
			description: "test underflow",
			input:       "9223372037s",
			error:       "underflow",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got, _, err := haproxytime.ParseDuration(tc.input)
			if err != nil && err.Error() != tc.error {
				t.Errorf("%q: wanted error %q, got %q", tc.input, tc.error, err)
				return
			}
			if got != tc.duration {
				t.Errorf("%q: wanted value %q, got %q", tc.input, tc.duration, got)
			}
		})
	}
}

func FuzzParseDuration(f *testing.F) {
	f.Add("")
	f.Add("0")
	f.Add("0d")
	f.Add("0ms")
	f.Add("1000garbage")
	f.Add("100us")
	f.Add("10s")
	f.Add("1d 3h")
	f.Add("1d")
	f.Add("1d3h30m45s")
	f.Add("1h30m")
	f.Add("5000")
	f.Add("500ms")
	f.Add("9223372036s")

	// Values extracted from the unit tests
	testCases := []string{
		"",
		"0",
		"a",
		"/",
		"100.d",
		"1d 30mgarbage",
		"1d 3h 30m 45s 100ms 200us",
		"1d3h30m45s100ms200us",
		"  1d   3h   30m   45s   ",
		"5000",
		"5000 100ms",
		"5000 100ms",
		"000000000000000000000001 01us",
		"0ms",
		"0d 0h 0m 0s 0ms 0us",
		"0d 0h 0m 0s 0 0us",
		"0d 0h 0m 0s 0ms 0",
		"0d 0h 0m 0s 1",
		"0ms 0ms",
		"1d 5m 1h",
		"1d 100us",
		"9223372036s",
		"9223372036s 1000ms",
		"9223372037s",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _, err := haproxytime.ParseDuration(input)
		if err != nil {
			t.Skip()
		}
	})
}
