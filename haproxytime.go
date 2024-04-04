package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/frobware/comptime"
)

// maxTimeout represents the maximum permissible timeout duration for
// HAProxy. Set at 2,147,483,647 milliseconds (approximately 24.8
// days), it aligns with the upper limit of HAProxy's timer
// configuration. This value corresponds to the maximum positive value
// for a signed 32-bit integer. Specifying a timeout exceeding this
// threshold (e.g., 2147483648ms) in HAProxy's configuration will
// result in an overflow error, causing a critical configuration
// failure, preventing HAProxy from starting. This constraint ensures
// that timeout values remain within the operational limits of
// HAProxy, regardless of the underlying system architecture.
const maxTimeout = 2147483647 * time.Millisecond

var (
	// buildVersion is a variable that should be populated at
	// build time using linker flags to specify the actual build
	// version. If it is not set, the default value "<unknown>"
	// will be used.
	buildVersion string = "<unknown>"
)

// version is a function that returns the build version.
func version() string {
	return buildVersion
}

var Usage = `
haproxytime - Convert human-readable time duration to millisecond format

General Usage:
  haproxytime [-help] [-v]
  haproxytime [-h] [-m] [<duration>]

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
  haproxytime -m           -> Print the maximum HAProxy duration.
  haproxytime 2h30m5s      -> Convert duration to milliseconds.
  haproxytime -h 4500000   -> Convert 4500000ms to a human-readable format.
  echo 150s | haproxytime  -> Convert 150 seconds to milliseconds.`[1:]

// ExitHandler defines an interface for handling exits.
type ExitHandler interface {
	Exit(code int)
}

// DefaultExitHandler is the production exit handler that calls
// os.Exit.
type DefaultExitHandler struct{}

func (e DefaultExitHandler) Exit(code int) {
	os.Exit(code)
}

// safeFprintf is a wrapper around fmt.Fprintf that performs a
// formatted write operation to a given io.Writer. It takes the same
// arguments as fmt.Fprintf: a format string and a variadic list of
// arguments. If the write operation fails, the function writes an
// error message to os.Stderr and exits the program using the provided
// ExitHandler.
func safeFprintf(w io.Writer, exitHandler ExitHandler, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error writing to output: %v\n", err)
		exitHandler.Exit(1)
	}
}

// safeFprintln is a wrapper around fmt.Fprintln that performs a write
// operation to a given io.Writer, appending a new line at the end. It
// takes the same variadic list of arguments as fmt.Fprintln. If the
// write operation fails, the function writes an error message to
// os.Stderr and exits the program using the provided ExitHandler.
func safeFprintln(w io.Writer, exitHandler ExitHandler, a ...interface{}) {
	_, err := fmt.Fprintln(w, a...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error writing to output: %v\n", err)
		exitHandler.Exit(1)
	}
}

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
func printErrorWithPosition(w io.Writer, exitHandler ExitHandler, input string, err error, position int) {
	safeFprintln(w, exitHandler, err)
	safeFprintln(w, exitHandler, input)
	safeFprintf(w, exitHandler, "%"+fmt.Sprint(position)+"s", "")
	safeFprintln(w, exitHandler, "^")
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
func output(w io.Writer, exitHandler ExitHandler, duration time.Duration, printHuman bool) {
	if printHuman {
		safeFprintln(w, exitHandler, formatDuration(duration))
	} else {
		safeFprintf(w, exitHandler, "%vms\n", duration.Milliseconds())
	}
}

// printPositionalError formats and outputs an error message to the
// provided io.Writer, along with the position at which the error
// occurred in the input argument. It supports error types with
// positional information, such as comptime.SyntaxError,
// comptime.OverflowError, and comptime.RangeError.
//
// Parameters:
//   - w: the io.Writer to output the error message, usually os.Stderr.
//   - exitHandler: the ExitHandler to handle exit scenarios.
//   - err: the error that occurred, expected to be of a type that
//     contains positional information, typically
//     *comptime.{OverflowError,RangeError,SyntaxError}.
//   - arg: the input argument string where the error occurred.
//
// The function uses the errors.As method to check if the error
// implements an interface that provides the Position() method. If
// such an error type is detected and is not nil, it prints the error
// message along with the position at which the error occurred, using
// the printErrorWithPosition function. If no matching error type is
// found, the function writes a generic error message and exits.
func printPositionalError(w io.Writer, exitHandler ExitHandler, err error, arg string) {
	var posErr interface {
		Position() int
	}
	if errors.As(err, &posErr) && posErr != nil {
		printErrorWithPosition(w, exitHandler, arg, err, posErr.Position())
		return
	}

	// Handle unexpected error types more gracefully.
	safeFprintf(w, exitHandler, "Unexpected error: %v\n", err)
	exitHandler.Exit(1)
}

// readAll reads all available bytes up to maxBytes from the given
// io.Reader into a string. It also trims any trailing newline
// characters. If an error occurs during the read operation, it
// returns an empty string and the error wrapped with additional
// context.
func readAll(rdr io.Reader, maxBytes int64) (string, error) {
	limitRdr := io.LimitReader(rdr, maxBytes)
	inputBytes, err := io.ReadAll(limitRdr)
	if err != nil {
		return "", fmt.Errorf("error reading: %w", err)
	}
	return strings.TrimRight(string(inputBytes), "\n"), nil
}

// readInput determines the source of the input for parsing the
// duration and retrieves the input. It first checks if there are any
// elements in the remainingArgs slice. If so, the first element of
// remainingArgs is returned. If remainingArgs is empty, the function
// reads from the provided io.Reader.
//
// Parameters:
//
//	rdr             An io.Reader from which to read if remainingArgs
//	                is empty.
//	remainingArgs   A slice of remaining command-line arguments.
//	maxBytes        The maximum number of bytes to read from the
//	                reader.
//
// Returns:
//
//	value           The first element of remainingArgs or the string
//	                read from the io.Reader.
//	err             An error if reading from the io.Reader fails.
func readInput(rdr io.Reader, remainingArgs []string, maxBytes int64) (string, error) {
	if len(remainingArgs) > 0 {
		return remainingArgs[0], nil
	}
	return readAll(rdr, maxBytes)
}

// convertDuration is the primary function for the haproxytime
// tool. It parses command-line flags, reads input for a duration
// string (either from arguments or stdin), converts it into a Go
// time.Duration object, and then outputs the result.
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
func convertDuration(rdr io.Reader, stdout, stderr io.Writer, args []string, exitHandler ExitHandler) int {
	fs := flag.NewFlagSet("haproxytime", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var showHelp, showVersion, printHuman, printMax bool

	fs.BoolVar(&printHuman, "h", false, "Print duration value in a human-readable format")
	fs.BoolVar(&printMax, "m", false, "Print the maximum HAProxy timeout value")
	fs.BoolVar(&showHelp, "help", false, "Show usage information")
	fs.BoolVar(&showVersion, "v", false, "Show version information")

	if err := fs.Parse(args); err != nil {
		safeFprintln(stderr, exitHandler, err)
		return 1
	}

	if showHelp {
		safeFprintln(stderr, exitHandler, Usage)
		return 1
	}

	if showVersion {
		safeFprintf(stderr, exitHandler, "haproxytime %s\n", version())
		return 0
	}

	if printMax {
		output(stdout, exitHandler, maxTimeout, printHuman)
		return 0
	}

	input, err := readInput(rdr, fs.Args(), 256)
	if err != nil {
		safeFprintln(stderr, exitHandler, err)
		return 1
	}

	duration, err := comptime.ParseDuration(input, comptime.Millisecond, comptime.ParseModeMultiUnit, func(position int, value time.Duration, totalSoFar time.Duration) bool {
		return value+totalSoFar <= maxTimeout
	})

	if err != nil {
		// If there are command-line arguments, print
		// positional error.
		if len(fs.Args()) > 0 {
			printPositionalError(stderr, exitHandler, err, fs.Args()[0])
			return 1
		}

		// If there are no command-line arguments, simply
		// print the error.
		safeFprintln(stderr, exitHandler, err)
		return 1
	}

	output(stdout, exitHandler, duration, printHuman)
	return 0
}

func main() {
	os.Exit(convertDuration(os.Stdin, os.Stdout, os.Stderr, os.Args[1:], DefaultExitHandler{}))
}
