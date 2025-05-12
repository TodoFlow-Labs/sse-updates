// internal/config/config.go
package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	HTTPAddr string `mapstructure:"http-addr"`
	NATSURL  string `mapstructure:"nats-url"`
	LogLevel string `mapstructure:"log-level"`
}

func Load() (*Config, error) {
	pflag.String("config", "config.yaml", "Path to config file")
	pflag.String("nats-url", "", "NATS server URL")
	pflag.String("log-level", "info", "Log level")
	pflag.String("http-addr", ":8081", "HTTP listen address")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, err
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigFile(viper.GetString("config"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("No config file found, using flags/env: %v\n", err)
	}

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.NATSURL == "" {
		return nil, fmt.Errorf("nats-url must be set")
	}
	if cfg.HTTPAddr == "" {
		return nil, fmt.Errorf("http-addr must be set")
	}

	// get environment from .env
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	return &cfg, nil
}
