package handler

import (
	"context"
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
)

// results count per page
var limit int64 = 10

// CreatePerson will handle the create person post request
func CreatePerson(db *mongo.Database, res http.ResponseWriter, req *http.Request) {
	var person model.Person
	json.NewDecoder(req.Body).Decode(&person)
	result, err := db.Collection("peopls").InsertOne(nil, person)
	if err != nil {
		log.Printf("Error while insert document: %v, type: %T\n", err, err)
		switch err.(type) {
		case mongo.WriteException:
			res.WriteHeader(http.StatusNotAcceptable)
			httpResponse := model.NewResponse(http.StatusNotAcceptable, "Error while inserting data.", nil)
			json.NewEncoder(res).Encode(httpResponse)
		default:
			res.WriteHeader(http.StatusInternalServerError)
			httpResponse := model.NewResponse(http.StatusInternalServerError, "Error while inserting data.", nil)
			json.NewEncoder(res).Encode(httpResponse)
		}
		return
	}
	res.WriteHeader(http.StatusCreated)
	person.ID = result.InsertedID.(primitive.ObjectID)
	httpResponse := model.NewResponse(http.StatusCreated, "", person)
	json.NewEncoder(res).Encode(httpResponse)
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
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := model.NewResponse(http.StatusInternalServerError, "Error happend while reading data", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	err = curser.All(context.Background(), &personList)
	if err != nil {
		log.Fatalf("Error in curser: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := model.NewResponse(http.StatusInternalServerError, "Error happend while reading data", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	httpResponse := model.NewResponse(http.StatusOK, "", personList)
	json.NewEncoder(res).Encode(httpResponse)
}

// GetPerson will give us person with special id
func GetPerson(db *mongo.Database, res http.ResponseWriter, req *http.Request) {
	var params = mux.Vars(req)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Printf("Error while decode from hex: %v", err)
		res.WriteHeader(http.StatusBadRequest)
		httpResponse := model.NewResponse(http.StatusBadRequest, "id that you sent is wrong!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	var person model.Person
	err = db.Collection("people").FindOne(nil, model.Person{ID: id}).Decode(&person)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			res.WriteHeader(http.StatusNotFound)
			httpResponse := model.NewResponse(http.StatusNotFound, "person not found", nil)
			json.NewEncoder(res).Encode(httpResponse)
			return
		}
		log.Printf("Error while decode to go struct:%v\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := model.NewResponse(http.StatusInternalServerError, "there is an error on server!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	res.WriteHeader(http.StatusOK)
	httpResponse := model.NewResponse(http.StatusOK, "", person)
	json.NewEncoder(res).Encode(httpResponse)
}

// UpdatePerson will handle the person update endpoint
func UpdatePerson(db *mongo.Database, res http.ResponseWriter, req *http.Request) {
	var updateData map[string]interface{}
	err := json.NewDecoder(req.Body).Decode(&updateData)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		httpResponse := model.NewResponse(http.StatusBadRequest, "json body is incorrect", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	// we dont handle the json decode return error because all our fields have the omitempty tag.
	var params = mux.Vars(req)
	oid, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Printf("Error while decode from hex: %v", err)
		res.WriteHeader(http.StatusBadRequest)
		httpResponse := model.NewResponse(http.StatusBadRequest, "id that you sent is wrong!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	update := bson.M{
		"$set": updateData,
	}
	result, err := db.Collection("people").UpdateOne(context.Background(), model.Person{ID: oid}, update)
	if err != nil {
		log.Printf("Error while updateing document: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		httpResponse := model.NewResponse(http.StatusInternalServerError, "error in updating document!!!", nil)
		json.NewEncoder(res).Encode(httpResponse)
		return
	}
	if result.MatchedCount == 1 {
		res.WriteHeader(http.StatusAccepted)
		httpResponse := model.NewResponse(http.StatusAccepted, "", &updateData)
		json.NewEncoder(res).Encode(httpResponse)
	} else {
		res.WriteHeader(http.StatusNotFound)
		httpResponse := model.NewResponse(http.StatusNotFound, "person not found", nil)
		json.NewEncoder(res).Encode(httpResponse)
	}
}
