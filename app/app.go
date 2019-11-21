package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/katoozi/golang-mongodb-rest-api/app/handler"
	"github.com/katoozi/golang-mongodb-rest-api/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// App has the mongo database and router instances
type App struct {
	Router *mux.Router
	DB     *mongo.Database
}

// Initialize initialize the app with
func (app *App) Initialize(config *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI()))
	if err != nil {
		log.Fatalf("Error while connecting to mongo: %v\n", err)
	}
	app.DB = client.Database("golang")
	app.setIndexes()

	app.Router = mux.NewRouter()
	app.Router.Use(handler.JSONContentTypeMiddleware)
	app.setRouters()
}

// SetupRouters will register routes in router
func (app *App) setRouters() {
	app.Post("/person", app.handleRequest(handler.CreatePerson))
	app.Patch("/person/{id}", app.handleRequest(handler.UpdatePerson))
	app.Put("/person/{id}", app.handleRequest(handler.UpdatePerson))
	app.Get("/person/{id}", app.handleRequest(handler.GetPerson))
	app.Get("/person", app.handleRequest(handler.GetPersons))
	app.Get("/person", app.handleRequest(handler.GetPersons), "page", "{page}")
}

// Get will register Get method for an endpoint
func (app *App) Get(path string, endpoint http.HandlerFunc, queries ...string) {
	app.Router.HandleFunc(path, endpoint).Methods("GET").Queries(queries...)
}

// Post will register Post method for an endpoint
func (app *App) Post(path string, endpoint http.HandlerFunc) {
	app.Router.HandleFunc(path, endpoint).Methods("POST")
}

// Put will register Put method for an endpoint
func (app *App) Put(path string, endpoint http.HandlerFunc) {
	app.Router.HandleFunc(path, endpoint).Methods("PUT")
}

// Patch will register Patch method for an endpoint
func (app *App) Patch(path string, endpoint http.HandlerFunc) {
	app.Router.HandleFunc(path, endpoint).Methods("PATCH")
}

// Delete will register Delete method for an endpoint
func (app *App) Delete(path string, endpoint http.HandlerFunc) {
	app.Router.HandleFunc(path, endpoint).Methods("DELETE")
}

// Run will start the http server on host that you pass in. host:<ip:port>
func (app *App) Run(host string) {
	// use signals for shutdown server gracefully.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt, os.Kill)
	go func() {
		log.Fatal(http.ListenAndServe(host, app.Router))
	}()
	log.Printf("Server is listning...")
	sig := <-sigs
	log.Println("Signal: ", sig)

	log.Println("Stoping MongoDB Connection...")
	app.DB.Client().Disconnect(context.Background())
}

// RequestHandlerFunction is a custome type that help us to pass db arg to all endpoints
type RequestHandlerFunction func(db *mongo.Database, w http.ResponseWriter, r *http.Request)

func (app *App) handleRequest(handler RequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(app.DB, w, r)
	}
}

// setIndexes will create unique and index fields.
func (app *App) setIndexes() {
	keys := bsonx.Doc{
		{Key: "username", Value: bsonx.Int32(1)},
		{Key: "email", Value: bsonx.Int32(1)},
	}
	index := mongo.IndexModel{}
	index.Keys = keys
	unique := true
	index.Options = &options.IndexOptions{
		Unique: &unique,
	}
	people := app.DB.Collection("people")
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := people.Indexes().CreateOne(context.Background(), index, opts)
	if err != nil {
		log.Fatalf("Error while creating indexs: %v", err)
	}
}
