# Parse composite time durations, including support for days

The `comptime` package (composite time) offers functionality for
parsing durations, extending the capabilities of the standard
library's `time.ParseDuration` function. It introduces support for an
additional time unit, 'days' (denoted by 'd'), and enables the parsing
of composite durations from a single string, such as '1d5m200ms'.

Key Features:

- Supports the following time units: "d" (days), "h" (hours), "m"
  (minutes), "s" (seconds), "ms" (milliseconds), and "us"
  (microseconds).
- Capable of parsing composite durations such as
  "24d20h31m23s647ms".
- Ensures parsed durations are non-negative.
- Custom Range Checking: Allows the caller to define their own range
  constraints on parsed durations through a BoundsChecker callback.
  This enables early termination of the parsing process based on
  user-defined limits.

## Dev Build

```sh
$ make
$ make benchmark
$ make benchmark-profile
```
