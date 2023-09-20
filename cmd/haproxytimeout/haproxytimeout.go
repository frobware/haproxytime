package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/frobware/haproxytime"
)

// These variable are populated at build time using linker flags, and
// the overall build version is retrieved via the Version function.
var (
	buildVersion string
)

// Version is a function variable that returns the current build
// version. By default, it returns the value of the unexported
// 'version' variable, which is set during build time. This variable
// is designed to be overridden for testing purposes.
var Version = func() string {
	return buildVersion
}

var Usage = `
haproxytimeout - Convert human-readable time durations to millisecond format

General Usage:
  haproxytimeout [-help] [-v]
  haproxytimeout [-h] [-m] [<duration>]

Usage:
  -help Show usage information
  -v	Show version information
  -h	Print duration value in a human-readable format
  -m	Print the maximum HAProxy timeout value
  <duration>: value to convert. If omitted, will read from stdin.

The flags [-help] and [-v] are mutually exclusive with any other
options or duration input.

Available units for time durations:
  d   days
  h:  hours
  m:  minutes
  s:  seconds
  ms: milliseconds
  us: microseconds

A duration value without a unit defaults to milliseconds.

Examples:
  haproxytimeout -m           -> Print the maximum HAProxy duration.
  haproxytimeout 2h30m5s      -> Convert duration to milliseconds.
  haproxytimeout -h 4500000   -> Convert 4500000ms to a human-readable format.
  echo 150s | haproxytimeout  -> Convert 150 seconds to milliseconds.`[1:]

// printErrorWithPosition writes an error message along with its
// position in the input string to the given Writer. The function
// prints the error, the input string, and a caret '^' pointing to the
// position where the error occurred.
//
// Parameters:
//   - w: the io.Writer to which the output is written
//   - input: the string that produced the error
//   - err: the error to be displayed
//   - position: the 1-based index at which the error occurred within the input
//
// Example:
//
//	If the input is "24d20h31m23s647msO000us" and the error
//	occurred at position 18, the output would be:
//
//	syntax error at position 18: invalid number
//	24d20h31m23s647msO000us
//			 ^
func printErrorWithPosition(w io.Writer, input string, err error, position int) {
	fmt.Fprintln(w, err)
	fmt.Fprintln(w, input)
	fmt.Fprintf(w, "%"+fmt.Sprint(position)+"s", "")
	fmt.Fprintln(w, "^")
}

// formatDuration takes a time.Duration value and returns a
// human-readable string representation. The string breaks down the
// duration into days, hours, minutes, seconds, milliseconds. Each
// unit of time will only be included in the output if its value is
// greater than zero.
//
// Example:
//
//	Input: 36h12m15s
//	Output: "1d12h12m15s"
//
//	Input: 2m15s300ms
//	Output: "2m15s300ms"
//
// Parameters:
//   - duration: the time.Duration value to be formatted
//
// Returns:
//   - A string representing the human-readable format of the input
//     duration.
func formatDuration(duration time.Duration) string {
	if duration == 0 {
		return "0ms"
	}

	const Day = time.Hour * 24
	days := duration / Day
	duration -= days * Day
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute
	duration -= minutes * time.Minute
	seconds := duration / time.Second
	duration -= seconds * time.Second
	milliseconds := duration / time.Millisecond

	var result string
	if days > 0 {
		result += fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm", minutes)
	}
	if seconds > 0 {
		result += fmt.Sprintf("%ds", seconds)
	}
	if milliseconds > 0 {
		result += fmt.Sprintf("%dms", milliseconds)
	}

	return result
}

// output writes a time.Duration value to the given io.Writer. The
// format of the output depends on the printHuman flag.
//
// Parameters:
//   - w: the io.Writer to which the output is written
//   - duration: the time.Duration value to be displayed
//   - printHuman: a boolean flag; if true display in human-readable format
//
// If printHuman is true, the duration is formatted using the
// formatDuration function, which breaks down the duration into units
// like days, hours, minutes, etc., and displays it accordingly.
//
// If printHuman is false, the duration is simply displayed as the
// number of milliseconds, followed by the unit "ms".
//
// Examples:
//   - With printHuman=true and duration=86400000ms, the output will be "1d".
//   - With printHuman=false and duration=86400000ms, the output will be "86400000ms".
func output(w io.Writer, duration time.Duration, printHuman bool) {
	if printHuman {
		fmt.Fprintln(w, formatDuration(duration))
	} else {
		fmt.Fprintf(w, "%vms\n", duration.Milliseconds())
	}
}

// printPositionalError formats and outputs an error message to the
// provided io.Writer, along with the position at which the error
// occurred in the input argument. It supports haproxytime.SyntaxError
// and haproxytime.OverflowError types, which contain positional
// information.
//
// Parameters:
//   - w: the io.Writer to output the error message, usually os.Stderr
//   - err: the error that occurred, expected to be of type *haproxytime.SyntaxError or *haproxytime.OverflowError
//   - arg: the input argument string where the error occurred
//
// The function first tries to cast the error to either
// haproxytime.SyntaxError or haproxytime.OverflowError. If
// successful, it prints the error message along with the position at
// which the error occurred, using printErrorWithPosition function.
//
// Example:
//
// Given an OverflowError with Position=18 and
// arg="24d20h31m23s647ms1000us", it would print:
//
//	  overflow error at position 18: value exceeds max duration
//	  24d20h31m23s647ms1000us
//			   ^
func printPositionalError(w io.Writer, err error, arg string) {
	var syntaxErr *haproxytime.SyntaxError
	var overflowErr *haproxytime.OverflowError

	switch {
	case errors.As(err, &syntaxErr):
		printErrorWithPosition(w, arg, err, syntaxErr.Position())
	case errors.As(err, &overflowErr):
		printErrorWithPosition(w, arg, err, overflowErr.Position())
	}
}

// readAll reads all available bytes from the given io.Reader into a
// string. It also trims any trailing newline characters. If an error
// occurs during the read operation, it returns an empty string and
// the error wrapped with additional context.
func readAll(rdr io.Reader) (string, error) {
	inputBytes, err := io.ReadAll(rdr)
	if err != nil {
		return "", fmt.Errorf("error reading: %w", err)
	}
	return strings.TrimRight(string(inputBytes), "\n"), nil
}

// getInputSource determines the source of the input for parsing the
// duration. It checks if there are any remaining command-line
// arguments, and if so, uses the first one as the input string.
// Otherwise, it reads from the provided Reader.
//
// Parameters:
//   - rdr: An io.Reader from which to read input if remainingArgs is
//     empty
//   - remainingArgs: A slice of remaining command-line arguments
//
// Returns:
//   - The input string to be parsed
//   - An error if reading from stdin fails
func getInputSource(rdr io.Reader, remainingArgs []string) (string, error) {
	if len(remainingArgs) > 0 {
		return remainingArgs[0], nil
	}
	return readAll(rdr)
}

// ConvertDuration is the primary logic function for the
// haproxytimeout tool. It parses command-line flags, reads input for
// a duration string (either from arguments or stdin), converts it
// into a Go time.Duration object, and then outputs the result.
//
// Parameters:
//   - stdin: the io.Reader from which input will be read.
//   - stdout: the io.Writer to which normal output will be written.
//   - stderr: the io.Writer to which error messages will be written.
//   - args: command-line arguments
//   - errorHandling: the flag.ErrorHandling strategy for parsing flags
//
// Returns:
//
//   - 0 for successful execution, 1 for errors
//
// Flags supported:
//   - help: Show usage information
//   - v: Show version information
//   - h: Output duration in a human-readable format
//   - m: Output the maximum HAProxy duration
//
// If an error occurs, the function writes the error message to stderr
// and returns 1. Otherwise, it writes the converted or maximum
// duration to stdout and returns 0.
func ConvertDuration(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("haproxytimeout", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var showHelp, showVersion, printHuman, printMax bool

	fs.BoolVar(&printHuman, "h", false, "Print duration value in a human-readable format")
	fs.BoolVar(&printMax, "m", false, "Print the maximum HAProxy timeout value")
	fs.BoolVar(&showHelp, "help", false, "Show usage information")
	fs.BoolVar(&showVersion, "v", false, "Show version information")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if showHelp {
		fmt.Fprintln(stderr, Usage)
		return 1
	}

	if showVersion {
		fmt.Fprintf(stderr, "haproxytimeout %s\n", Version())
		return 0
	}

	if printMax {
		output(stdout, haproxytime.MaxTimeout, printHuman)
		return 0
	}

	input, err := getInputSource(stdin, fs.Args())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	duration, err := haproxytime.ParseDuration(input, haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit)
	if err != nil {
		if len(fs.Args()) > 0 {
			printPositionalError(stderr, err, fs.Args()[0])
		} else {
			fmt.Fprintln(stderr, err)
		}
		return 1
	}

	output(stdout, duration, printHuman)
	return 0
}

func main() {
	os.Exit(ConvertDuration(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]))
}