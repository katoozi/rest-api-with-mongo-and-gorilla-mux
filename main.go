package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// Person is the data structure that we will save and receive.
type Person struct {
	ID        primitive.ObjectID     `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string                 `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName  string                 `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Username  string                 `json:"username,omitempty" bson:"username,omitempty"`
	Email     string                 `json:"email,omitempty" bson:"email,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`
}

var people *mongo.Collection

// results count per page
var limit int64 = 10

func createIndexModels() mongo.IndexModel {
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
	return index
}

func main() {
	// read environment variables
	user := os.Getenv("mongo_user")
	password := os.Getenv("mongo_password")
	host := os.Getenv("mongo_host")
	port := os.Getenv("mongo_port")
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", user, password, host, port)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Error while connecting to mongo: %v\n", err)
	}

	people = client.Database("golang").Collection("people")
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	index := createIndexModels()
	_, err = people.Indexes().CreateOne(context.Background(), index, opts)
	if err != nil {
		log.Fatalf("Error while creating indexs: %v", err)
	}

	router := mux.NewRouter()
	router.Use(JSONContentTypeMiddleware)
	router.HandleFunc("/person", CreatePerson).Methods("POST")
	router.HandleFunc("/person/{id}", UpdatePerson).Methods("PATCH")
	router.HandleFunc("/person/{id}", UpdatePerson).Methods("PUT")
	router.HandleFunc("/person", GetPersons).Methods("GET").Queries("page", "{page}")
	router.HandleFunc("/person", GetPersons).Methods("GET")
	router.HandleFunc("/person/{id}", GetPerson).Methods("GET")

	fmt.Println("Server is listening...")
	http.ListenAndServe(":1234", router)
}

// CreatePerson will handle the create person post request
func CreatePerson(res http.ResponseWriter, req *http.Request) {
	var person Person
	json.NewDecoder(req.Body).Decode(&person)
	result, err := people.InsertOne(nil, person)
	if err != nil {
		log.Printf("Error while insert document: %v, type: %T\n", err, err)
		switch err.(type) {
		case mongo.WriteException:
			res.WriteHeader(http.StatusNotAcceptable)
			httpResponse := response(http.StatusNotAcceptable, "Error while inserting data.", nil)
			json.NewEncoder(res).Encode(httpResponse)
		default:
			res.WriteHeader(http.StatusInternalServerError)
			httpResponse := response(http.StatusInternalServerError, "Error while inserting data.", nil)
			json.NewEncoder(res).Encode(httpResponse)
		}
		return
	}
	res.WriteHeader(http.StatusCreated)
	person.ID = result.InsertedID.(primitive.ObjectID)
	httpResponse := response(http.StatusCreated, "", person)
	json.NewEncoder(res).Encode(httpResponse)
}

// GetPersons will handle people list get request
func GetPersons(res http.ResponseWriter, req *http.Request) {
	var personList []Person
	pageString := req.FormValue("page")
	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		page = 0
	}
	page = page * limit
	findOptions := options.FindOptions{
		Skip:  &page,
		Limit: &limit,
		Sort: bson.M{
			"_id": -1, // -1 for descending and 1 for ascending
		},
	}
	curser, err := people.Find(nil, bson.M{}, &findOptions)
	if err != nil {
		log.Printf("Error while quering collection: %v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := response(http.StatusInternalServerError, "Error happend while reading data", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	err = curser.All(context.Background(), &personList)
	if err != nil {
		log.Fatalf("Error in curser: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := response(http.StatusInternalServerError, "Error happend while reading data", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	httpResponse := response(http.StatusOK, "", personList)
	json.NewEncoder(res).Encode(httpResponse)
}

// GetPerson will give us person with special id
func GetPerson(res http.ResponseWriter, req *http.Request) {
	var params = mux.Vars(req)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Printf("Error while decode from hex: %v", err)
		res.WriteHeader(http.StatusBadRequest)
		httpResponse := response(http.StatusBadRequest, "id that you sent is wrong!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	var person Person
	err = people.FindOne(nil, Person{ID: id}).Decode(&person)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			res.WriteHeader(http.StatusNotFound)
			httpResponse := response(http.StatusNotFound, "person not found", nil)
			json.NewEncoder(res).Encode(httpResponse)
			return
		}
		log.Printf("Error while decode to go struct:%v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := response(http.StatusInternalServerError, "there is an error on server!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	res.WriteHeader(http.StatusOK)
	httpResponse := response(http.StatusOK, "", person)
	json.NewEncoder(res).Encode(httpResponse)
}

// UpdatePerson will handle the person update endpoint
func UpdatePerson(res http.ResponseWriter, req *http.Request) {
	var person Person
	json.NewDecoder(req.Body).Decode(&person)
	// we dont handle the json decode return error because all our fields have the omitempty tag.
	var params = mux.Vars(req)
	oid, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Printf("Error while decode from hex: %v", err)
		res.WriteHeader(http.StatusBadRequest)
		httpResponse := response(http.StatusBadRequest, "id that you sent is wrong!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	update := bson.M{
		"$set": person,
	}
	result, err := people.UpdateOne(context.Background(), Person{ID: oid}, update)
	if err != nil {
		log.Printf("Error while updateing document: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := response(http.StatusInternalServerError, "error in updating document!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	res.WriteHeader(http.StatusAccepted)
	httpResponse := response(http.StatusAccepted, "", result)
	json.NewEncoder(res).Encode(httpResponse)
}
