package haproxytime_test

import (
	"fmt"
	"os"

	"github.com/frobware/haproxytime"
)

// ExampleParseDuration gets copied to the documentation.
func ExampleParseDuration() {
	if duration, err := haproxytime.ParseDuration("1h", haproxytime.UnitMillisecond, haproxytime.ParseModeSingleUnit); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("%vms\n", duration.Milliseconds())
	}

	if duration, err := haproxytime.ParseDuration("24d20h31m23s647ms", haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("%vms\n", duration.Milliseconds())
	}

	if duration, err := haproxytime.ParseDuration("500", haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("%vms\n", duration.Milliseconds())
	}

	// Output:
	// 3600000ms
	// 2147483647ms
	// 500ms

}
