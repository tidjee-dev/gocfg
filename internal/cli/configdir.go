package cli

import (
	"fmt"

	"github.com/tidjee-dev/gocfg/internal/scaffold"
)

// legacyConfigDir is the pre-v2 layout, honored with a warning.
const legacyConfigDir = "config"

// checkConfigDir verifies the configuration directory. An explicitly
// chosen directory is strict: a missing dir is an error even when the
// legacy layout exists. The default falls back to legacy "config/"
// with a warning, and always reports it when both exist.
func checkConfigDir(dir string, explicit bool, warn func(string)) error {
	if isDir(dir) {
		if !explicit && dir == scaffold.DefaultConfigDir && isDir(legacyConfigDir) {
			warn("legacy config/ found, move to internal/config")
		}
		return nil
	}
	if !explicit && dir == scaffold.DefaultConfigDir && isDir(legacyConfigDir) {
		warn("legacy config/ found, move to internal/config")
		return nil
	}
	return fmt.Errorf("%s/ not found (run gocfg init)", dir)
}
