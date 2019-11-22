package main

import "github.com/katoozi/golang-mongodb-rest-api/config"

import "github.com/katoozi/golang-mongodb-rest-api/app"

import "log"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := config.NewConfig()
	app := app.NewApp(config)
	app.Run(":1234")
}
