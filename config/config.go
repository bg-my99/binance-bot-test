package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AccessKeys struct {
		ApiKey    string `envconfig:"API_KEY"`
		SecretKey string `envconfig:"SECRET_KEY"`
	}
	UseTestnet   bool   `yaml:"useTestNet"`
	TradesSource string `yaml:"tradesSource"`
	WriteTrades  bool   `yaml:"writeTrades"`
	FetchForDate string `envconfig:"FETCH_DATE"`
}

func ReadEnv(cfg *Config) {

	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Println(err)
	}

	err = envconfig.Process("", cfg)
	if err != nil {
		fmt.Println(err)
	}
}
