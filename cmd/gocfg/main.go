package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/tidjee-dev/gocfg/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		var ee *cli.ExitError
		if errors.As(err, &ee) {
			fmt.Fprintln(os.Stderr, "Error:", ee.Err)
			os.Exit(ee.Code) // 1 = configuration/validation failure
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2) // 2 = CLI usage error
	}
}
