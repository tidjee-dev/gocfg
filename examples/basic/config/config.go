package config

import "fmt"

// Config is the composed application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

// Load resolves every configuration section. Parsing errors are returned,
// never hidden behind fallbacks.
func Load() (Config, error) {
	app, err := App()
	if err != nil {
		return Config{}, err
	}
	db, err := Database()
	if err != nil {
		return Config{}, err
	}
	return Config{App: app, Database: db}, nil
}

// Validate enforces domain rules on top of parsing.
func (c Config) Validate() error {
	switch c.App.Env {
	case "dev", "test", "prod":
		return nil
	default:
		return fmt.Errorf("unsupported APP_ENV %q: expected dev, test or prod", c.App.Env)
	}
}
