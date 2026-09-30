package dotenv

import (
	"os"
)

// Load parses paths (default ".env" if none) and sets only missing OS vars.
// Enforces OS > .env. Missing file with default path is a no-op (nil).
func Load(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{".env"}
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) && len(paths) == 1 && p == ".env" {
				return nil
			}
			return err
		}
		m, err := Parse(data)
		if err != nil {
			return err
		}
		for k, v := range m {
			if _, ok := os.LookupEnv(k); !ok {
				_ = os.Setenv(k, v)
			}
		}
	}
	return nil
}
