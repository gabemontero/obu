package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/gabemontero/obu/pkg/cmd/cli"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	command := cli.CommandFor()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "obu encountered the following error: %v\n", err)
		os.Exit(1)
	}
}
