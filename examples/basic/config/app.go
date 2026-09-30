package config

import "github.com/tidjee-dev/gocfg/env"

// AppConfig holds the application-level configuration.
type AppConfig struct {
	Name  string
	Env   string
	Debug bool
	URL   string
}

// App resolves the application configuration.
func App() (AppConfig, error) {
	name, err := env.String("APP_NAME", "My App")
	if err != nil {
		return AppConfig{}, err
	}
	envVal, err := env.String("APP_ENV", "dev")
	if err != nil {
		return AppConfig{}, err
	}
	debug, err := env.Bool("APP_DEBUG", true)
	if err != nil {
		return AppConfig{}, err
	}
	url, err := env.String("APP_URL", "http://localhost:9000")
	if err != nil {
		return AppConfig{}, err
	}
	return AppConfig{Name: name, Env: envVal, Debug: debug, URL: url}, nil
}
