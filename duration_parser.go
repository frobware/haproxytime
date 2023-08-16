// Package haproxytime provides specialized duration parsing
// functionality with features beyond the standard library's
// time.ParseDuration function. It adds support for extended time
// units such as "days", denoted by "d", and optionally allows the
// parsing of multi-unit durations in a single string like
// "1d5m200ms".
//
// Key Features:
//
//   - Supports the following time units: "d" (days), "h" (hours), "m"
//     (minutes), "s" (seconds), "ms" (milliseconds), and "us"
//     (microseconds).
//
//   - Capable of parsing composite durations such as
//     "24d20h31m23s647ms".
//
//   - Ensures parsed durations are non-negative.
//
//   - Respects HAProxy's maximum duration limit of 2147483647ms.
package haproxytime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// These constants represent different units of time used in the
// duration parsing process. They are ordered in decreasing magnitude,
// from UnitDay to UnitMicrosecond. The zero value of the unit type is
// reserved to represent an invalid unit.
const (
	UnitDay Unit = iota + 1
	UnitHour
	UnitMinute
	UnitSecond
	UnitMillisecond
	UnitMicrosecond
)

const (
	// ParseModeMultiUnit allows for multiple units to be
	// specified together in the duration string, e.g., "1d2h3m".
	ParseModeMultiUnit ParseMode = iota + 1

	// ParseModeSingleUnit permits only a single unit type to be
	// present in the duration string. Any subsequent unit types
	// will result in an error. For instance, "1d" would be valid,
	// but "1d2h" would not.
	ParseModeSingleUnit

	// MaxTimeout represents the maximum timeout duration,
	// equivalent to the maximum signed 32-bit integer value in
	// milliseconds.
	MaxTimeout = 2147483647 * time.Millisecond
)

// ParseMode defines the behavior for interpreting units in a duration
// string. It decides how many units can be accepted when parsing.
type ParseMode int

// Unit is used to represent different time units (day, hour, minute,
// second, millisecond, microsecond) in numerical form. The zero value
// represents an invalid time unit.
type Unit uint

// unitInfo defines a time unit's symbol and its corresponding
// duration in time.Duration units.
type unitInfo struct {
	// symbol is the string representation of the time unit, e.g.,
	// "h" for hour.
	symbol string

	// duration represents the length of time that one unit of
	// this type equates to, measured in Go's time.duration units.
	duration time.Duration
}

// SyntaxError represents an error that occurs during the parsing of a
// duration string. It provides details about the specific nature of
// the error and the position in the string where the error was
// detected.
type SyntaxError struct {
	// cause specifies the type of syntax error encountered, such
	// as InvalidNumber, InvalidUnit, InvalidUnitOrder, or
	// UnexpectedCharactersInSingleUnitMode.
	cause SyntaxErrorCause

	// position represents the location in the input string where
	// the error was detected. The position is 0-indexed.
	position int
}

// SyntaxErrorCause represents the cause of a syntax error during
// duration parsing. It discriminates between different kinds of
// syntax errors to aid in error handling and debugging.
type SyntaxErrorCause int

const (
	// InvalidNumber indicates that a provided number in the
	// duration string is invalid or cannot be interpreted.
	InvalidNumber SyntaxErrorCause = iota + 1

	// InvalidUnit signifies that an unrecognised or unsupported
	// unit is used in the duration string.
	InvalidUnit

	// InvalidUnitOrder denotes an error when units in the
	// duration string are not in decreasing order of magnitude
	// (e.g., specifying minutes before hours).
	InvalidUnitOrder

	// UnexpectedCharactersInSingleUnitMode indicates that
	// unexpected characters were encountered beyond the first
	// valid duration when parsing in ParseModeSingleUnit. This
	// occurs when multiple unit-value pairs or extraneous
	// characters are found, which are not permitted in this mode.
	UnexpectedCharactersInSingleUnitMode
)

// OverflowError represents an error that occurs when a parsed value
// exceeds the allowable range, leading to an overflow condition.
type OverflowError struct {
	// position represents the location in the input string where
	// the error was detected. The position is 0-indexed.
	position int

	// number is the substring of the input string that represents
	// the numeric value causing the overflow. This provides a
	// direct reference to the original representation of the
	// number in the input.
	number string
}

var (
	// unitProperties maps Units to their details.
	unitProperties = map[Unit]unitInfo{
		UnitDay:         {symbol: "d", duration: 24 * time.Hour},
		UnitHour:        {symbol: "h", duration: time.Hour},
		UnitMinute:      {symbol: "m", duration: time.Minute},
		UnitSecond:      {symbol: "s", duration: time.Second},
		UnitMillisecond: {symbol: "ms", duration: time.Millisecond},
		UnitMicrosecond: {symbol: "us", duration: time.Microsecond},
	}

	// symbolToUnit maps time unit symbols to their corresponding
	// Units.
	symbolToUnit = map[string]Unit{
		"d":  UnitDay,
		"h":  UnitHour,
		"m":  UnitMinute,
		"s":  UnitSecond,
		"ms": UnitMillisecond,
		"us": UnitMicrosecond,
	}
)

// consumeUnit scans the input string starting from the given position
// and attempts to extract a known time unit symbol. It first looks
// for multi-character symbols like "ms" and "us". If none of the
// multi-character symbols are found, it returns the single character
// at the current position as the consumed unit symbol.
//
// This function is exclusively called by ParseDuration; it is never
// called when there is no remaining input.
//
// Parameters:
//   - input: The string being parsed.
//   - start: The starting position for scanning the string.
//
// Returns:
//   - A string representation of the found unit symbol.
//   - The new position in the string after the last character of the unit symbol.
func consumeUnit(input string, start int) (string, int) {
	current := start
	for _, symbol := range []string{"ms", "us"} {
		if strings.HasPrefix(input[current:], symbol) {
			return symbol, current + len(symbol)
		}
	}
	return string(input[current]), current + 1
}

// consumeNumber scans the input string starting from the given
// position and attempts to extract a contiguous sequence of numeric
// characters (digits).
//
// This function is exclusively called by ParseDuration; it is never
// called when there is no remaining input.
//
// Parameters:
//   - input: The string being parsed.
//   - start: The starting position for scanning the string.
//
// Returns:
//   - A string representation of the contiguous sequence of digits.
//   - The new position in the string after the last digit.
func consumeNumber(input string, start int) (string, int) {
	current := start
	for current < len(input) && unicode.IsDigit(rune(input[current])) {
		current++
	}
	if start == current {
		return "", current
	}
	return input[start:current], current
}

// Is checks whether the provided target error matches the SyntaxError
// type. This method facilitates the use of the errors.Is function for
// matching against SyntaxError.
//
// Example:
//
//	if errors.Is(err, &haproxytime.SyntaxError{}) {
//	    // handle SyntaxError
//	}
func (e *SyntaxError) Is(target error) bool {
	_, ok := target.(*SyntaxError)
	return ok
}

// Position returns the position in the input string where the
// SyntaxError occurred. The position is 0-based, meaning that the
// first character in the input string is at position 0.
func (e *SyntaxError) Position() int {
	return e.position
}

// Error implements the error interface for ParseError. It provides a
// formatted error message detailing the position and the nature of
// the parsing error. Note that the position is reported as 1-index
// based.
func (e *SyntaxError) Error() string {
	var msg string
	switch e.cause {
	case InvalidNumber:
		msg = "invalid number"
	case InvalidUnit:
		msg = "invalid unit"
	case InvalidUnitOrder:
		msg = "invalid unit order"
	case UnexpectedCharactersInSingleUnitMode:
		msg = "unexpected characters in single unit mode"
	}
	return fmt.Sprintf("syntax error at position %d: %v", e.position+1, msg)
}

// Cause returns the specific cause of the SyntaxError. The cause
// provides details on the type of syntax error encountered, such as
// InvalidNumber, InvalidUnit, InvalidUnitOrder, or
// UnexpectedCharactersInSingleUnitMode.
func (e *SyntaxError) Cause() SyntaxErrorCause {
	return e.cause
}

// Is checks whether the provided target error matches the
// OverflowError type. This method facilitates the use of the
// errors.Is function for matching against OverflowError.
//
// Example:
//
//	if errors.Is(err, &haproxytime.OverflowError{}) {
//	    // handle OverflowError
//	}
func (e *OverflowError) Is(target error) bool {
	_, ok := target.(*OverflowError)
	return ok
}

// Position returns the position in the input string where the
// OverflowError occurred. The position is 0-based, indicating that
// the first character in the input string is at position 0.
func (e *OverflowError) Position() int {
	return e.position
}

// Error returns a formatted message indicating the position and value
// that caused the overflow, and includes additional context from any
// underlying error, if present. The position is reported as
// 1-indexed.
func (e *OverflowError) Error() string {
	return fmt.Sprintf("overflow error at position %v: value exceeds max duration", e.position+1)
}

// newOverflowError creates a new OverflowError instance. position
// specifies the 0-indexed position in the input string where the
// overflow error was detected. number is the numeric value in string
// form that caused the overflow.
func newOverflowError(position int, number string) *OverflowError {
	return &OverflowError{
		position: position,
		number:   number,
	}
}

// newSyntaxErrorInvalidNumber creates a new SyntaxError instance with
// the InvalidNumber cause. position specifies the 0-indexed position
// in the input string where the invalid number was detected.
func newSyntaxErrorInvalidNumber(position int) *SyntaxError {
	return &SyntaxError{
		cause:    InvalidNumber,
		position: position,
	}
}

// newSyntaxErrorInvalidUnit creates a new SyntaxError instance with
// the InvalidUnit cause. position specifies the 0-indexed position in
// the input string where the invalid unit was detected.
func newSyntaxErrorInvalidUnit(position int) *SyntaxError {
	return &SyntaxError{
		cause:    InvalidUnit,
		position: position,
	}
}

// newSyntaxErrorInvalidUnitOrder creates a new SyntaxError instance
// with the InvalidUnitOrder cause. position specifies the 0-indexed
// position in the input string where the invalid unit order was
// detected.
func newSyntaxErrorInvalidUnitOrder(position int) *SyntaxError {
	return &SyntaxError{
		cause:    InvalidUnitOrder,
		position: position,
	}
}

// newSyntaxErrorUnexpectedCharactersInSingleUnitMode creates a new
// SyntaxError instance with the UnexpectedCharactersInSingleUnitMode
// cause. position specifies the 0-indexed position in the input
// string where the extraneous characters were detected.
func newSyntaxErrorUnexpectedCharactersInSingleUnitMode(position int) *SyntaxError {
	return &SyntaxError{
		cause:    UnexpectedCharactersInSingleUnitMode,
		position: position,
	}
}

// ParseDuration translates an input string representing a time
// duration into a time.Duration type. The string may include values
// with the following units: "d" (days), "h" (hours), "m" (minutes),
// "s" (seconds), "ms" (milliseconds), "us" (microseconds).
//
// Input examples:
//   - 10s
//   - 1h30m
//   - 500ms
//   - 100us
//   - 1d5m200
//   - 1000
//
// The last two examples both contain values (e.g., 200 and 1000) that
// lack a unit specifier. These values will be interpreted according
// to the default unit provided as an argument to the ParseDuration
// function.
//
// An empty input results in a zero duration.
//
// Returns a time.Duration representing the parsed duration value from
// the input string. If the input is invalid or cannot be parsed into
// a valid time.Duration, the function will return one of the two
// following error types:
//
//   - SyntaxError: When the input has non-numeric values,
//     unrecognised units, improperly formatted values, or units that
//     are not in descending order from day to microsecond.
//
//   - OverflowError: If the total duration exceeds HAProxy's maximum
//     limit or any individual value in the input leads to an overflow
//     in the total duration.
func ParseDuration(input string, defaultUnit Unit, parseMode ParseMode) (time.Duration, error) {
	position := 0 // in input

	var newTotal time.Duration
	var prevUnit Unit

	for position < len(input) {
		numberPosition := position
		numberStr, newPos := consumeNumber(input, position)
		position = newPos
		if numberStr == "" {
			return 0, newSyntaxErrorInvalidNumber(numberPosition)
		}

		value, err := strconv.ParseInt(numberStr, 10, 32)
		if err != nil {
			return 0, newOverflowError(numberPosition, numberStr)
		}

		unitPosition := position
		var unitStr string
		if position < len(input) {
			unitStr, newPos = consumeUnit(input, position)
			position = newPos
		} else {
			unitStr = unitProperties[defaultUnit].symbol
		}

		unit, ok := symbolToUnit[unitStr]
		if !ok {
			return 0, newSyntaxErrorInvalidUnit(unitPosition)
		}

		if prevUnit != 0 && unit <= prevUnit {
			return 0, newSyntaxErrorInvalidUnitOrder(unitPosition)
		}

		segmentDuration := time.Duration(value) * unitProperties[unit].duration
		if newTotal > MaxTimeout-segmentDuration {
			return 0, newOverflowError(numberPosition, numberStr)
		}

		newTotal += segmentDuration
		prevUnit = unit

		if parseMode == ParseModeSingleUnit && position < len(input) {
			return 0, newSyntaxErrorUnexpectedCharactersInSingleUnitMode(position)
		}
	}

	return newTotal, nil
}
