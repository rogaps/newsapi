package newsapi

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
	"os"
)

// Config represents server configurations
type Config struct {
	Server struct {
		Address string `json:"address" env:"SERVER_ADDRESS" envDefault:":8080"`
	} `json:"server"`
	Postgres struct {
		ConnectionString string `json:"connectionString" env:"POSTGRES_CONNECTION_STRING"`
	} `json:"postgres"`
	Elasticsearch struct {
		ConnectionString []string `json:"connectionString" env:"ES_CONNECTION_STRING" envSeparator:";"`
	} `json:"elasticsearch"`
	Redis struct {
		ConnectionString string `json:"connectionString" env:"REDIS_CONNECTION_STRING"`
	} `json:"redis"`
	AMQP struct {
		ConnectionString string `json:"connectionString" env:"AMQP_CONNECTION_STRING"`
	} `json:"amqp"`
}

// Parse parses config from config file
func (config *Config) Parse(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("parse config failed: %s", err)
	}
	if err = json.NewDecoder(file).Decode(config); err != nil {
		return fmt.Errorf("parse config failed: %s", err)
	}
	return nil
}

func NewConfig(fileName string) func() *Config {
	return func() *Config {
		config := Config{}
		if err := env.Parse(&config); err != nil {
			log.Errorln(err)
		}
		if len(fileName) > 0 {
			if err := config.Parse(fileName); err != nil {
				log.Errorln(err)
			}
		}
		return &config
	}
}
