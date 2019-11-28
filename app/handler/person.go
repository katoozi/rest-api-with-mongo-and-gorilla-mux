package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/katoozi/golang-mongodb-rest-api/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// results count per page
var limit int64 = 10

// CreatePerson will handle the create person post request
func CreatePerson(db *mongo.Database, res http.ResponseWriter, req *http.Request) {
	person := new(model.Person)
	err := json.NewDecoder(req.Body).Decode(person)
	if err != nil {
		ResponseWriter(res, http.StatusBadRequest, "body json request have issues!!!", nil)
		return
	}
	result, err := db.Collection("people").InsertOne(nil, person)
	if err != nil {
		switch err.(type) {
		case mongo.WriteException:
			ResponseWriter(res, http.StatusNotAcceptable, "username or email already exists in database.", nil)
		default:
			ResponseWriter(res, http.StatusInternalServerError, "Error while inserting data.", nil)
		}
		return
	}
	person.ID = result.InsertedID.(primitive.ObjectID)
	ResponseWriter(res, http.StatusCreated, "", person)
}

// GetPersons will handle people list get request
func GetPersons(db *mongo.Database, res http.ResponseWriter, req *http.Request) {
	var personList []model.Person
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
	curser, err := db.Collection("people").Find(nil, bson.M{}, &findOptions)
	if err != nil {
		log.Printf("Error while quering collection: %v\n", err)
		ResponseWriter(res, http.StatusInternalServerError, "Error happend while reading data", nil)
		return
	}
	err = curser.All(context.Background(), &personList)
	if err != nil {
		log.Fatalf("Error in curser: %v", err)
		ResponseWriter(res, http.StatusInternalServerError, "Error happend while reading data", nil)
		return
	}
	ResponseWriter(res, http.StatusOK, "", personList)
}

// GetPerson will give us person with special id
func GetPerson(db *mongo.Database, res http.ResponseWriter, req *http.Request) {
	var params = mux.Vars(req)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		ResponseWriter(res, http.StatusBadRequest, "id that you sent is wrong!!!", nil)
		return
	}
	var person model.Person
	err = db.Collection("people").FindOne(nil, model.Person{ID: id}).Decode(&person)
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			ResponseWriter(res, http.StatusNotFound, "person not found", nil)
		default:
			log.Printf("Error while decode to go struct:%v\n", err)
			ResponseWriter(res, http.StatusInternalServerError, "there is an error on server!!!", nil)
		}
		return
	}
	ResponseWriter(res, http.StatusOK, "", person)
}

// UpdatePerson will handle the person update endpoint
func UpdatePerson(db *mongo.Database, res http.ResponseWriter, req *http.Request) {
	var updateData map[string]interface{}
	err := json.NewDecoder(req.Body).Decode(&updateData)
	if err != nil {
		ResponseWriter(res, http.StatusBadRequest, "json body is incorrect", nil)
		return
	}
	// we dont handle the json decode return error because all our fields have the omitempty tag.
	var params = mux.Vars(req)
	oid, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		ResponseWriter(res, http.StatusBadRequest, "id that you sent is wrong!!!", nil)
		return
	}
	update := bson.M{
		"$set": updateData,
	}
	result, err := db.Collection("people").UpdateOne(context.Background(), model.Person{ID: oid}, update)
	if err != nil {
		log.Printf("Error while updateing document: %v", err)
		ResponseWriter(res, http.StatusInternalServerError, "error in updating document!!!", nil)
		return
	}
	if result.MatchedCount == 1 {
		ResponseWriter(res, http.StatusAccepted, "", &updateData)
	} else {
		ResponseWriter(res, http.StatusNotFound, "person not found", nil)
	}
}
