package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AccessKeys struct {
		ApiKey    string `envconfig:"API_KEY"`
		SecretKey string `envconfig:"SECRET_KEY"`
	}
}

func ReadEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		fmt.Println(err)
	}
}
