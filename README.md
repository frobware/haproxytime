# parse time durations, with support for days

This is a Go library that parses duration strings in a format similar
to time.ParseDuration, with the additional capability of handling
duration strings specifying a number of days ("d"). This functionality
is not available in the built-in time.ParseDuration function. It also
differs by not accepting negative values. This package was primarily
created for validating HAProxy timeout values.

The CLI utility `haproxy-timeout-checker` is an example of using the
package. It validates the time duration using `ParseDuration` and also
checks to see if the duration exceeds HAProxy's maximum.

```console
$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "9223372036s"
duration 9223372036000ms exceeds HAProxy's maximum value of 2147483647ms

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "2147483647ms"
2147483647

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "2147483648ms"
duration 2147483648ms exceeds HAProxy's maximum value of 2147483647ms

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "1d"
86400000

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "1d 1s"
86401000

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "1d 3h 10m 20s 100ms 9999us"
97820109

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go 5000
5000

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "5000 999999ms"
5000 999999ms
            ^
error: invalid unit order

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "1d 1f"
1d 1f
    ^
error: invalid unit

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "1d 1d"
1d 1d
    ^
error: invalid unit order

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "1d 5m 1230ms"
86701230

# Note: Spaces are optional.
$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "1d5m"
86700000

$ go run cmd/haproxy-timeout-checker/haproxy-timeout-checker.go "9223372037s"
9223372037s
          ^
error: underflow
```
