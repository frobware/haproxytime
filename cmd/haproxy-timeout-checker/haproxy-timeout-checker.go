package main

import (
	"fmt"
	"os"
	"time"

	"github.com/frobware/haproxytime"
)

const HAProxyMaxTimeoutValue = 2147483647 * time.Millisecond

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`usage: <duration>`)
		os.Exit(1)
	}

	duration, position, err := haproxytime.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, os.Args[1])
		fmt.Fprintf(os.Stderr, "%"+fmt.Sprint(position)+"s", "")
		fmt.Fprintln(os.Stderr, "^")
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if duration.Milliseconds() > HAProxyMaxTimeoutValue.Milliseconds() {
		fmt.Fprintf(os.Stderr, "duration %vms exceeds HAProxy's maximum value of %vms\n", duration.Milliseconds(), HAProxyMaxTimeoutValue.Milliseconds())
		os.Exit(1)
	}

	fmt.Println(duration.Milliseconds())
}
