// Package comptime (composite time) offers functionality for parsing
// durations, extending the capabilities of the standard library's
// `time.ParseDuration` function. It introduces support for an
// additional time unit, 'days' (denoted by 'd'), and enables the
// parsing of composite durations from a single string, such as
// '1d5m200ms'.
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
//   - Custom Range Checking: Allows the user to define their own range
//     constraints on parsed durations through a BoundsChecker callback.
//     This enables early termination of the parsing process based on
//     user-defined limits.
package comptime

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// These constants represent different units of time used in the
// duration parsing process. They are ordered in increasing order of
// magnitude, from Microsecond to Day.
const (
	Microsecond Unit = iota
	Millisecond
	Second
	Minute
	Hour
	Day
)

const (
	// ParseModeMultiUnit allows for multiple units to be
	// specified together in the duration string, e.g., "1d2h3m".
	ParseModeMultiUnit ParseMode = iota

	// ParseModeSingleUnit permits only a single unit type to be
	// present in the duration string. Any subsequent unit types
	// will result in an error. For instance, "1d" would be valid,
	// but "1d2h" would not.
	ParseModeSingleUnit
)

// ParseMode defines the behavior for interpreting units in a duration
// string. It decides how many units can be accepted when parsing.
type ParseMode int

// Unit is used to represent different time units (day, hour, minute,
// second, millisecond, microsecond) in numerical form. The zero value
// represents an invalid time unit.
type Unit uint

// SyntaxError represents an error that occurs during the parsing of a
// duration string. It provides details about the specific nature of
// the error and the position in the string where the error was
// detected.
type SyntaxError struct {
	// cause specifies the type of syntax error encountered, such
	// as InvalidNumber, InvalidUnit, InvalidUnitOrder, or
	// UnexpectedCharactersInSingleUnitMode.
	cause SyntaxErrorCause

	// position represents the 0-based index location in the input
	// string where the condition for a SyntaxError was triggered.
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
	// position represents the 0-based index location in the input
	// string where the condition for a OverflowError was
	// triggered.
	position int
}

// RangeError represents a condition where a parsed value exceeds a
// user-defined allowable range.
type RangeError struct {
	// position represents the 0-based index location in the input
	// string where the condition for a RangeError was triggered.
	position int
}

// unitDuration consolidates a time unit and its respective duration.
type unitDuration struct {
	// unit represents the time unit as defined by the Unit
	// enumeration.
	unit Unit

	// duration specifies the duration one unit represents,
	// measured in time.Duration.
	duration time.Duration
}

// unitProperties provides constant-time access to Unit enumeration
// values and their properties. The order of values in unitProperties
// should match the order of values in the Unit enumeration for
// consistency.
var unitProperties = [6]unitDuration{
	{Microsecond, time.Microsecond},
	{Millisecond, time.Millisecond},
	{Second, time.Second},
	{Minute, time.Minute},
	{Hour, time.Hour},
	{Day, 24 * time.Hour},
}

// consumeUnit scans the input string starting from the given position
// and attempts to extract a known time unit symbol. It first looks
// for multi-character symbols like "ms" and "us". If none of the
// multi-character symbols are found, it checks for single-character
// units like "h", "m", "s", and "d". If a valid unit is found, it
// returns true along with the corresponding Unit enum value. If no
// valid unit is found, it returns false.
//
// This function is exclusively called by ParseDuration; it is never
// called when there is no remaining input.
//
// Parameters:
//   - input: The string being parsed.
//   - start: The starting position for scanning the string.
//
// Returns:
//   - A Unit enum value representing the found unit if valid.
//   - The new position in the string after the last character of the unit symbol.
//   - A bool indicating whether a valid Unit was matched.
func consumeUnit(input string, start int) (Unit, int, bool) {
	if len(input) > start+1 && input[start+1] == 's' {
		switch input[start] {
		case 'm':
			return Millisecond, start + 2, true
		case 'u':
			return Microsecond, start + 2, true
		}
	}

	switch input[start] {
	case 'h':
		return Hour, start + 1, true
	case 'm':
		return Minute, start + 1, true
	case 's':
		return Second, start + 1, true
	case 'd':
		return Day, start + 1, true
	default:
		// Must return a Unit, so we return UnitDay, but false
		// takes precedence (i.e., no known unit was matched).
		return Day, start, false
	}
}

// consumeNumberError represents error codes for parsing numbers in
// the input string.
type consumeNumberError int

const (
	// noNumberFound indicates that no numeric characters were
	// found.
	noNumberFound consumeNumberError = iota + 1

	// overflow indicates that an overflow occurred while parsing
	// the number.
	overflow
)

// consumeNumber scans the input string starting from the given
// position and attempts to extract a contiguous sequence of numeric
// characters (digits).
//
// Parameters:
//   - input: The string being parsed.
//   - start: The starting position for scanning the string.
//
// Returns:
//
//   - The parsed integer value.
//
//   - The new position in the string after the last digit.
//
//   - A consumeNumberError indicating whether no number was found or
//     if an overflow occurred.
func consumeNumber(input string, start int) (int64, int, consumeNumberError) {
	const maxInt64Div10 = math.MaxInt64 / 10

	var value int64
	position := start

	for position < len(input) {
		c := input[position]
		if c >= '0' && c <= '9' {
			digit := int64(c - '0')
			if value > maxInt64Div10 || (value == maxInt64Div10 && digit > 7) {
				return 0, position, overflow
			}
			value = value*10 + digit
		} else {
			break
		}
		position += 1
	}

	if position == start {
		return 0, position, noNumberFound
	}

	return value, position, 0
}

// Is checks whether the provided target error matches the SyntaxError
// type. This method facilitates the use of the errors.Is function for
// matching against SyntaxError.
//
// Example:
//
//	if errors.Is(err, &comptime.SyntaxError{}) {
//	    // handle SyntaxError
//	}
func (e *SyntaxError) Is(target error) bool {
	var syntaxError *SyntaxError
	ok := errors.As(target, &syntaxError)
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
// the parsing error. The position in the error message is converted
// to 1-based indexing, rather than the original 0-based indexing used
// in the input string.
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

// Is checks whether the provided target error matches the RangeError
// type. This method facilitates the use of the errors.Is function for
// matching against RangeError.
//
// Example:
//
//	if errors.Is(err, &comptime.RangeError{}) {
//	    // handle RangeError
//	}
func (e *RangeError) Is(target error) bool {
	var rangeError *RangeError
	ok := errors.As(target, &rangeError)
	return ok
}

// Position returns the position in the input string where the
// RangeError occurred. The position is 0-based, meaning that the
// first character in the input string is at position 0.
func (e *RangeError) Position() int {
	return e.position
}

// Is checks whether the provided target error matches the
// OverflowError type. This method facilitates the use of the
// errors.Is function for matching against OverflowError.
//
// Example:
//
//	if errors.Is(err, &comptime.OverflowError{}) {
//	    // handle OverflowError
//	}
func (e *OverflowError) Is(target error) bool {
	var overflowError *OverflowError
	ok := errors.As(target, &overflowError)
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
// underlying error, if present. The position in the error message is
// converted to 1-based indexing, rather than the original 0-based
// indexing used in the input string. 1-indexed.
func (e *OverflowError) Error() string {
	return fmt.Sprintf("overflow error at position %v", e.position+1)
}

// Error returns a formatted message indicating the position where the
// allowable range was exceeded. The position in the error message is
// converted to 1-based indexing, rather than the original 0-based
// indexing used in the input string.
func (e *RangeError) Error() string {
	return fmt.Sprintf("range error at position %d", e.position+1)
}

// newOverflowError creates a new OverflowError instance. position
// specifies the 0-based index in the input string where the overflow
// occurs.
func newOverflowError(position int) *OverflowError {
	return &OverflowError{
		position: position,
	}
}

// newRangeError creates a new RangeError instance. position specifies
// the 0-based index in the input string where the range error was
// triggered.
func newRangeError(position int) *RangeError {
	return &RangeError{
		position: position,
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

// RangeChecker is a function type that serves as a callback during
// the parsing process in ParseDuration. The callback is invoked every
// time a new composite duration (unit * value) is calculated.
//
// Parameters:
//
//   - position: The current position in the input string where the
//     composite duration was parsed.
//
//   - value: The composite duration that was just calculated
//     (unit * value).
//
//   - totalSoFar: The total duration that has been parsed so far.
//
// The callback returns a boolean that determines whether parsing
// should continue (true) or stop immediately (false). If the parsing
// is halted, a RangeError will be returned from the ParseDuration
// function.
type RangeChecker func(position int, value time.Duration, totalSoFar time.Duration) bool

// NoRangeChecking is a sentinel BoundsChecker function that allows
// the ParseDuration function to proceed without any range checks.
// This function always returns true, allowing the parsing to
// continue.
func NoRangeChecking(position int, value time.Duration, totalSoFar time.Duration) bool {
	return true
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
// a valid time.Duration, the function will return one of the
// following error types:
//
//   - SyntaxError: When the input has non-numeric values,
//     unrecognised units, improperly formatted values, or units that
//     are not in descending order from day to microsecond.
//
//   - OverflowError: If the total duration exceeds the maximum
//     limit that can be represented as a time.Duration, or if any
//     individual value in the input leads to an overflow in the
//     total duration. Takes precedence over RangeError.
//
//   - RangeError: If the parsing is halted by a BoundsChecker callback
//     returning false.
//
// Note: If both OverflowError and RangeError conditions are met,
// OverflowError will take precedence.
func ParseDuration(input string, defaultUnit Unit, parseMode ParseMode, inRangeChecker RangeChecker) (time.Duration, error) {
	position := 0 // in input
	var totalDuration time.Duration
	var prevUnit = Day

	for position < len(input) {
		numStartPos := position
		value, numEndPos, parseNumErr := consumeNumber(input, numStartPos)

		switch parseNumErr {
		case noNumberFound:
			return 0, newSyntaxErrorInvalidNumber(numStartPos)
		case overflow:
			return 0, newOverflowError(numStartPos)
		}

		var unit = defaultUnit
		var unitEndPos int
		var unitStartPos = numEndPos

		if unitStartPos < len(input) {
			var validUnit bool
			unit, unitEndPos, validUnit = consumeUnit(input, unitStartPos)
			if !validUnit {
				return 0, newSyntaxErrorInvalidUnit(unitStartPos)
			}
		}

		if position > 0 && unit >= prevUnit {
			return 0, newSyntaxErrorInvalidUnitOrder(unitStartPos)
		}
		prevUnit = unit

		compositeDuration := time.Duration(value) * unitProperties[unit].duration
		if compositeDuration < 0 || totalDuration > (math.MaxInt64-compositeDuration) {
			return 0, newOverflowError(numStartPos)
		}

		if inRangeChecker != nil && !inRangeChecker(position, compositeDuration, totalDuration) {
			return 0, newRangeError(numStartPos)
		}

		totalDuration += compositeDuration

		// Update position before error checking to reflect
		// the extent of parsing.
		position = max(unitEndPos, numEndPos)

		// In single-unit mode, check for unexpected
		// characters using the updated position. This helps
		// accurately pinpoint where in the input string the
		// unexpected characters start, indicating a syntax
		// error if additional characters are present.
		if parseMode == ParseModeSingleUnit && position < len(input) {
			return 0, newSyntaxErrorUnexpectedCharactersInSingleUnitMode(position)
		}
	}

	return totalDuration, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
