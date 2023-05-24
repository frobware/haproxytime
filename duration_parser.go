// Package haproxytime provides a utility for parsing
// duration strings in a format similar to time.ParseDuration, with
// the additional capability of handling duration strings specifying a
// number of days (d). This functionality is not available in the
// built-in time.ParseDuration function. It also returns an error if
// any duration is negative.
//
// This package was primarily created for validating HAProxy timeout
// values.
//
// For example, an input of "2d 4h" would be parsed into a
// time.Duration representing two days and four hours.
package haproxytime

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode"
)

var (
	// errParseDurationOverflow is triggered when a duration value
	// exceeds the permissible maximum limit, leading to an
	// overflow.
	errParseDurationOverflow = fmt.Errorf("overflow")

	// errParseDurationUnderflow is triggered when a duration
	// value falls below the acceptable minimum limit, resulting
	// in an underflow.
	errParseDurationUnderflow = fmt.Errorf("underflow")

	// strToUnit is a map that associates string representations
	// of time units with their corresponding unit constants. It
	// serves as a lookup table to convert string units to their
	// respective durations.
	strToUnit = map[string]unit{
		"d":  day,
		"h":  hour,
		"m":  minute,
		"s":  second,
		"ms": millisecond,
		"us": microsecond,
	}

	// unitToDuration is a map that correlates time duration units
	// with their corresponding durations in time.Duration format.
	unitToDuration = map[unit]time.Duration{
		day:         24 * time.Hour,
		hour:        time.Hour,
		minute:      time.Minute,
		second:      time.Second,
		millisecond: time.Millisecond,
		microsecond: time.Microsecond,
	}
)

// unit is used to represent different time units (day, hour, minute,
// second, millisecond, microsecond) in numerical form. The zero value
// of 'unit' represents an invalid time unit.
type unit uint

// These constants represent different units of time used in the
// duration parsing process. They are ordered in decreasing magnitude
// from day to microsecond. The zero value of 'unit' type is reserved
// to represent an invalid unit.
const (
	day unit = iota + 1
	hour
	minute
	second
	millisecond
	microsecond
)

type token struct {
	value int64
	unit
	duration time.Duration
}

type parser struct {
	tokens   []*token
	current  int // current offset in input
	position int // parse error location in input
}

func (p *parser) parse(input string) error {
	p.tokens = make([]*token, 0, len(input)/2)
	p.skipWhitespace(input)

	for p.current < len(input) {
		p.position = p.current
		token, err := p.nextToken(input)
		if err != nil {
			return err
		}
		if err := p.validateToken(token); err != nil {
			return err
		}
		p.tokens = append(p.tokens, token)
		p.skipWhitespace(input)
	}

	return nil
}

func (p *parser) validateToken(token *token) error {
	if len(p.tokens) > 0 {
		prevUnit := p.tokens[len(p.tokens)-1].unit
		if prevUnit >= token.unit {
			return fmt.Errorf("invalid unit order")
		}
	}
	if token.duration < 0 {
		return errParseDurationUnderflow
	}
	return nil
}

func (p *parser) nextToken(input string) (*token, error) {
	p.position = p.current
	value, err := p.consumeNumber(input)
	if err != nil {
		return nil, err
	}

	p.position = p.current
	unitStr := p.consumeUnit(input)
	if unitStr == "" {
		unitStr = "ms"
	}

	unit, found := strToUnit[unitStr]
	if !found {
		return nil, errors.New("invalid unit")
	}

	return &token{value, unit, time.Duration(value) * unitToDuration[unit]}, nil
}

func (p *parser) consumeNumber(input string) (int64, error) {
	start := p.current
	for p.current < len(input) && unicode.IsDigit(rune(input[p.current])) {
		p.current++
	}
	if start == p.current {
		// This yields a better error message compared to what
		// strconv.ParseInt returns for the empty string.
		return 0, errors.New("invalid number")
	}
	return strconv.ParseInt(input[start:p.current], 10, 64)
}

func (p *parser) consumeUnit(input string) string {
	start := p.current
	for p.current < len(input) && unicode.IsLetter(rune(input[p.current])) {
		p.current++
	}
	return input[start:p.current]
}

func (p *parser) skipWhitespace(input string) {
	for p.current < len(input) && unicode.IsSpace(rune(input[p.current])) {
		p.current++
	}
}

// ParseDuration translates a string representing a time duration into
// a time.Duration type. The input string can comprise duration values
// with units of days ("d"), hours ("h"), minutes ("m"), seconds
// ("s"), milliseconds ("ms"), and microseconds ("us"). If no unit is
// provided, the default is milliseconds ("ms"). If the input string
// comprises multiple duration values, they are summed to calculate
// the total duration. For example, "1h30m" is interpreted as 1 hour +
// 30 minutes.
//
// Returns:
//
//   - A time.Duration value representing the total duration found in
//     the string.
//
//   - An integer value indicating the position in the input string
//     where parsing failed.
//
//   - An error value representing any parsing error that occurred.
//
// Errors:
//
//   - It returns an "invalid number" error when a non-numeric or
//     improperly formatted numeric value is found.
//
//   - It returns an "invalid unit" error when an unrecognised or
//     invalid time unit is provided.
//
//   - It returns an "invalid unit order" error when the time units in
//     the input string are not in descending order from day to
//     microsecond or when the same unit is specified more than once.
//
//   - It returns an "overflow" error when the total duration value
//     exceeds the maximum possible value that a time.Duration can
//     represent.
//
//   - It returns an "underflow" error if any individual time value in
//     the input cannot be represented by time.Duration. For example,
//     a duration of "9223372036s 1000ms" would return an underflow
//     error.
//
// The function extracts duration values and their corresponding units
// from the input string and calculates the total duration. It
// tolerates missing units as long as the subsequent units are
// presented in descending order of "d", "h", "m", "s", "ms", and
// "us". A duration value provided without a unit is treated as
// milliseconds by default.
//
// Some examples of valid input strings are: "10s", "1h 30m", "500ms",
// "100us", "1d 5m 200"; in the last example 200 will default to
// milliseconds. Spaces are also optional.
//
// If an empty string is given as input, the function returns zero for
// the duration and no error.
func ParseDuration(input string) (time.Duration, int, error) {
	p := parser{}

	if err := p.parse(input); err != nil {
		return 0, p.position, err
	}

	checkedAddDurations := func(x, y time.Duration) (time.Duration, error) {
		result := x + y
		if x > 0 && y > 0 && result < 0 {
			return 0, errParseDurationOverflow
		}
		return result, nil
	}

	var err error
	var total time.Duration

	for i := range p.tokens {
		if total, err = checkedAddDurations(total, p.tokens[i].duration); err != nil {
			return 0, 0, err
		}
	}

	return total, 0, nil
}
