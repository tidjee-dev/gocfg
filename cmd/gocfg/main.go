// Command gocfg manages Go application configuration: initialization,
// environment files, and validation. Configuration itself stays plain
// Go code (see github.com/tidjee-dev/gocfg); this CLI only handles
// project files.
//
// Install:
//
//	go install github.com/tidjee-dev/gocfg/cmd/gocfg@latest
//
// Commands:
//
//	gocfg init      create .env, .env.example and config/
//	gocfg env       complete .env from .env.example (append-missing-only)
//	gocfg check     static health inspection (files, parse, key presence)
//	gocfg validate  validate values (schema presence or typed specs)
//	gocfg version   print the version
//
// Exit codes: 0 success, 1 configuration/validation failure,
// 2 CLI usage error.
//
// Release builds should inject the version:
//
//	go build -ldflags "-X github.com/tidjee-dev/gocfg/internal/cli.Version=v0.1.0" ./cmd/gocfg
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/tidjee-dev/gocfg/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if ee, ok := errors.AsType[*cli.ExitError](err); ok {
			fmt.Fprintln(os.Stderr, "Error:", ee.Err)
			os.Exit(ee.Code) // 1 = configuration/validation failure
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2) // 2 = CLI usage error
	}
}
