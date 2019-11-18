package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	// Person is the data structure that we will save and receive.
	Person struct {
		ID        primitive.ObjectID     `json:"_id,omitempty" bson:"_id,omitempty"`
		FirstName string                 `json:"first_name,omitempty" bson:"first_name,omitempty"`
		LastName  string                 `json:"last_name,omitempty" bson:"last_name,omitempty"`
		Username  string                 `json:"username,omitempty" bson:"username,omitempty"`
		Email     string                 `json:"email,omitempty" bson:"email,omitempty"`
		Data      map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`
	}

	// Response is the http json response schema
	Response struct {
		Status  int         `json:"status"`
		Message string      `json:"message"`
		Content interface{} `json:"content"`
	}

	// PaginatedResponse is the paginated response json schema
	PaginatedResponse struct {
		Count    int         `json:"count"`
		Next     string      `json:"next"`
		Previous string      `json:"previous"`
		Results  interface{} `json:"results"`
	}
)

func response(status int, message string, content interface{}) *Response {
	return &Response{
		Status:  status,
		Message: message,
		Content: content,
	}
}
