package config

import (
	"fmt"
	"os"
)

// Config is the server configuration structure
type Config struct {
	// mongo
	MongoUser     string
	MongoPassword string
	MongoHost     string
	MongoPort     string
}

// Initialize will read env variables and save them in config structure
func (config *Config) Initialize() {
	// read environment variables
	config.MongoUser = os.Getenv("mongo_user")
	config.MongoPassword = os.Getenv("mongo_password")
	config.MongoHost = os.Getenv("mongo_host")
	config.MongoPort = os.Getenv("mongo_port")
}

// MongoURI will generate mongo db connect uri
func (config *Config) MongoURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s",
		config.MongoUser,
		config.MongoPassword,
		config.MongoHost,
		config.MongoPort,
	)
}

// NewConfig will create and initialize config struct
func NewConfig() *Config {
	config := Config{}
	config.Initialize()
	return &config
}
