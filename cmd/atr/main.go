package main

import (
	"errors"
	"fmt"
	"os"

	"atr/internal/atcoder"
)

const version = "0.3.0"

// errUsage marks errors caused by wrong invocation (exit 2, not 1).
var errUsage = errors.New("usage error")

const usage = `Usage:
  atr new|n [-s] <contest ID (e.g. abc300)>       set up a contest (-s: select tasks)
  atr download|d <URL or problem ID (e.g. abc300_a)>
  atr test|t [options]   run tests (see: atr test -h)
`

func main() {
	atcoder.UserAgent = "atr/" + version
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "new", "n":
		err = cmdNew(os.Args[2:])
	case "download", "d":
		err = cmdDownload(os.Args[2:])
	case "test", "t":
		err = cmdTest(os.Args[2:])
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		if errors.Is(err, errUsage) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
