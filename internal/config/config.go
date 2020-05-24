package config

import "time"

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel        int           `mapstructure:"log_level"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	}

	PostgreSQL struct {
		DSN                string `mapstructure:"dsn"`
		MaxOpenConnections int    `mapstructure:"max_open_connections"`
		MaxIdleConnections int    `mapstructure:"max_idle_connections"`
		Automigrate        bool   `mapstructure:"automigrate"`
	} `mapstructure:"postgresql"`

	Integration struct {
		DSN string `mapstructure:"dsn"`
	}
}

// C holds the global configuration.
var C Config
