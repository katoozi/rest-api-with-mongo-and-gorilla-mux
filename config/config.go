package config

import (
	"fmt"
	"os"
)

// Config is the server configuration structure.
// all fields will be filled with environment variables.
type Config struct {
	ServerHost    string // address that server will listening on
	MongoUser     string // mongo db username
	MongoPassword string // mongo db password
	MongoHost     string // host that mongo db listening on
	MongoPort     string // port that mongo db listening on
}

// Initialize will read environment variables and save them in config structure fields
func (config *Config) Initialize() {
	// read environment variables
	config.ServerHost = os.Getenv("server_host")
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
