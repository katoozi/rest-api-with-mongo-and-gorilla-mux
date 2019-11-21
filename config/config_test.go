package config

import "testing"

import "os"

import "fmt"

const succeed = "\u2713"
const failed = "\u2717"

func TestConfig(t *testing.T) {
	os.Setenv("server_host", ":1234")
	os.Setenv("mongo_user", "john")
	os.Setenv("mongo_password", "doe")
	os.Setenv("mongo_host", "localhost")
	os.Setenv("mongo_port", "27017")

	config := NewConfig()

	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		"john",
		"doe",
		"localhost",
		"27017",
	)

	if config.MongoURI() != mongoURI {
		t.Fatalf("%s there is an problem in mongo connection uri generator. %s != %s", failed, config.MongoURI(), mongoURI)
	}
	t.Logf("%s Testing MongoDB connection uri generator is successful", succeed)
}
