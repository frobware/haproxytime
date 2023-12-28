# Parse time durations, with support for days

Package haproxytime provides specialised duration parsing
functionality with features beyond the standard library's
time.ParseDuration function. It adds support for extended time units
such as "days", denoted by "d", and optionally allows the parsing of
composite durations in a single string like "1d5m200ms".

Key Features:

- Supports the following time units: "d" (days), "h" (hours), "m"
  (minutes), "s" (seconds), "ms" (milliseconds), and "us"
  (microseconds).
- Capable of parsing composite durations such as
  "24d20h31m23s647ms".
- Ensures parsed durations are non-negative.

The command line utility `haproxytimeout` is an example of using the
package but also serves to convert human-readable duration values to
microseconds, suitable for a HAProxy configuration file.

## Install

To install with versioning information

```sh
$ go install -ldflags "-X 'main.buildVersion=$(git describe --tags --abbrev=8 --dirty --always --long)'" github.com/frobware/haproxytime/cmd/haproxytimeout@latest
```

To install without versioning information

```sh
$ go install github.com/frobware/haproxytime/cmd/haproxytimeout@latest
```

## Usage

```console
haproxytimeout - Convert human-readable time duration to millisecond format

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
  echo 150s | haproxytimeout  -> Convert 150 seconds to milliseconds.

```

## Dev Build

```sh
$ make
```
