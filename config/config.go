package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Env                     string
	HTTP_SERVER_PORT        string `envconfig:"HTTP_SERVER_PORT" default:"3001"`
	GIN_MODE                string `envconfig:"GIN_MODE" default:"release"`
	MYSQL_CONNECTION_STRING string `envconfig:"MYSQL_CONNECTION_STRING"`
}

func Load() Config {
	var config Config
	ENV, ok := os.LookupEnv("ENV")
	if !ok {
		// Default value for ENV.
		ENV = "dev"
	}
	// Load the .env file only for dev env.
	ENV_CONFIG, ok := os.LookupEnv("ENV_CONFIG")
	if !ok {
		ENV_CONFIG = "./.env"
	}

	err := godotenv.Load(ENV_CONFIG)
	if err != nil {
		logrus.Warn("Can't load env file")
	}

	envconfig.MustProcess("", &config)
	config.Env = ENV

	return config
}
