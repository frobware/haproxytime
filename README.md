# haproxytime

`haproxytime` is a command-line utility designed to facilitate the
conversion of human-readable time durations into a format precisely
represented in milliseconds, tailored specifically for HAProxy
configurations.

This tool is useful for developers and administrators who work with
HAProxy, providing a straightforward mechanism to ensure that timeout
configurations are accurately specified in a format that HAProxy can
understand. Whether you're inputting durations directly or piping them
from standard input, `haproxytime` streamlines the process of dealing
with time durations, supporting a wide range of units from days to
microseconds. With features to display durations in both machine- and
human-readable formats, as well as the ability to print the maximum
HAProxy timeout value, `haproxytime` minimises the risk of
configuration errors in your HAProxy setup.

## Install

To install without versioning information:

```sh
$ go install github.com/frobware/haproxytime@latest
```

To install with versioning information:

```sh
$ go install -ldflags "-X 'main.buildVersion=$(git describe --tags --abbrev=8 --dirty --always --long)'" github.com/frobware/haproxytime@latest
```

## Usage

```console
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
  echo 150s | haproxytime  -> Convert 150 seconds to milliseconds.
```

## Build

```sh
$ make
$ make install
```
