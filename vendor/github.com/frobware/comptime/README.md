# Parse time durations, with support for days

Package comptime provides specialised duration parsing
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

## Dev Build

```sh
$ make
$ make benchmark
$ make benchmark-profile
```
