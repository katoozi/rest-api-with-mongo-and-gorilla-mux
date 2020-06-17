package main

import (
	"log"

	"github.com/katoozi/golang-mongodb-rest-api/app"
	"github.com/katoozi/golang-mongodb-rest-api/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := config.NewConfig()
	app.ConfigAndRunApp(config)
}
