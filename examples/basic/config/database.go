package config

import "github.com/tidjee-dev/gocfg/env"

// DatabaseConfig holds the database configuration.
type DatabaseConfig struct {
	URL      string
	MaxConns int
}

// Database resolves the database configuration.
func Database() (DatabaseConfig, error) {
	url, err := env.String("DATABASE_URL", "postgres://localhost:5432/myapp")
	if err != nil {
		return DatabaseConfig{}, err
	}
	maxConns, err := env.Int("DATABASE_MAX_CONNS", 10)
	if err != nil {
		return DatabaseConfig{}, err
	}
	return DatabaseConfig{URL: url, MaxConns: maxConns}, nil
}
