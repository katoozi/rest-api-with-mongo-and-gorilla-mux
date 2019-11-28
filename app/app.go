package app

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/net/context"
	"github.com/gorilla/mux"
	"github.com/katoozi/golang-mongodb-rest-api/app/db"
	"github.com/katoozi/golang-mongodb-rest-api/app/handler"
	"github.com/katoozi/golang-mongodb-rest-api/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// App has the mongo database and router instances
type App struct {
	Router *mux.Router
	DB     *mongo.Database
}

// ConfigAndRunApp will create and initialize App structure. App factory function.
func ConfigAndRunApp(config *config.Config) {
	app := new(App)
	app.Initialize(config)
	app.Run(config.ServerHost)
}

// Initialize initialize the app with
func (app *App) Initialize(config *config.Config) {
	app.DB = db.InitialConnection("golang", config.MongoURI())
	app.createIndexes()

	app.Router = mux.NewRouter()
	app.UseMiddleware(handler.JSONContentTypeMiddleware)
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

// UseMiddleware will add global middleware in router
func (app *App) UseMiddleware(middleware mux.MiddlewareFunc) {
	app.Router.Use(middleware)
}

// createIndexes will create unique and index fields.
func (app *App) createIndexes() {
	// username and email will be unique.
	keys := bsonx.Doc{
		{Key: "username", Value: bsonx.Int32(1)},
		{Key: "email", Value: bsonx.Int32(1)},
	}
	people := app.DB.Collection("people")
	db.SetIndexes(people, keys)
}

// Get will register Get method for an endpoint
func (app *App) Get(path string, endpoint http.HandlerFunc, queries ...string) {
	app.Router.HandleFunc(path, endpoint).Methods("GET").Queries(queries...)
}

// Post will register Post method for an endpoint
func (app *App) Post(path string, endpoint http.HandlerFunc, queries ...string) {
	app.Router.HandleFunc(path, endpoint).Methods("POST").Queries(queries...)
}

// Put will register Put method for an endpoint
func (app *App) Put(path string, endpoint http.HandlerFunc, queries ...string) {
	app.Router.HandleFunc(path, endpoint).Methods("PUT").Queries(queries...)
}

// Patch will register Patch method for an endpoint
func (app *App) Patch(path string, endpoint http.HandlerFunc, queries ...string) {
	app.Router.HandleFunc(path, endpoint).Methods("PATCH").Queries(queries...)
}

// Delete will register Delete method for an endpoint
func (app *App) Delete(path string, endpoint http.HandlerFunc, queries ...string) {
	app.Router.HandleFunc(path, endpoint).Methods("DELETE").Queries(queries...)
}

// Run will start the http server on host that you pass in. host:<ip:port>
func (app *App) Run(host string) {
	// use signals for shutdown server gracefully.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt, os.Kill)
	go func() {
		log.Fatal(http.ListenAndServe(host, app.Router))
	}()
	log.Printf("Server is listning on http://%s\n", host)
	sig := <-sigs
	log.Println("Signal: ", sig)

	log.Println("Stoping MongoDB Connection...")
	app.DB.Client().Disconnect(context.Background())
}

// RequestHandlerFunction is a custome type that help us to pass db arg to all endpoints
type RequestHandlerFunction func(db *mongo.Database, w http.ResponseWriter, r *http.Request)

// handleRequest is a middleware we create for pass in db connection to endpoints.
func (app *App) handleRequest(handler RequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(app.DB, w, r)
	}
}
