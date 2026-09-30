// Command gocfg manages Go application configuration:
// initialization, environment files, and validation.
//
// Release builds should inject the version:
//
//	go build -ldflags "-X github.com/tidjee-dev/gocfg/internal/cli.Version=v0.1.0" ./cmd/gocfg
package main

import (
	"os"

	"github.com/tidjee-dev/gocfg/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
