package config

import "github.com/tidjee-dev/gocfg/env"

// Application variables mirror App()/Database() as explicit definitions:
// the single source of truth the CLI consumes (see Definitions).
var (
	AppName          = env.StringVar("APP_NAME", "My App")
	AppEnv           = env.StringVar("APP_ENV", "dev")
	AppDebug         = env.BoolVar("APP_DEBUG", true)
	AppURL           = env.URLVar("APP_URL", "http://localhost:9000")
	DatabaseURL      = env.URLVar("DATABASE_URL", "postgres://localhost:5432/myapp")
	DatabaseMaxConns = env.IntVar("DATABASE_MAX_CONNS", 10)
)

// Definitions exposes every configuration variable for tooling
// (`gocfg env --defs`, `gocfg validate --defs`). Keep it in sync
// with App() and Database().
func Definitions() []env.Any {
	return []env.Any{
		AppName,
		AppEnv,
		AppDebug,
		AppURL,
		DatabaseURL,
		DatabaseMaxConns,
	}
}
