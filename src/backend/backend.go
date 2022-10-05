package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.TODO()

type User struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Email     string             `bson:"email"`
	Name      string             `bson:"name"`
	//Nation    Nation             `bson:"nation"`
}

type Nation struct {
	ID    primitive.ObjectID `bson:"_id"`
	Name  string             `bson:"name"`
	Group string             `bson:"group"`
}

func createUser(user *User) error {
	_, err := collection.InsertOne(ctx, user)
	return err
}

func requestHandler(w http.ResponseWriter, req *http.Request) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{}

	if err != nil {
		fmt.Println(err.Error())
	}

	collection = client.Database("sweepstakes").Collection("users")

	data := map[string]interface{}{}

	err = json.NewDecoder(req.Body).Decode(&data)

	if err != nil {
		fmt.Println(err.Error())
	}

	switch req.Method {
	case "POST":
		response, err = createRecord(collection, ctx, data)
	case "GET":
		response, err = getRecords(collection, ctx)
		//case "PUT":
		//    response, err = updateRecord(collection, ctx, data)
		//case "DELETE":
		//    response, err = deleteRecord(collection, ctx, data)
	}

	if err != nil {
		response = map[string]interface{}{"error": err.Error()}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(response); err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	router := mux.NewRouter()

	// Handle API routes
	api := router.PathPrefix("/api/").Subrouter()

	// Respond to requests made to /api/users
	api.HandleFunc("/users", requestHandler)

	// Serve static files

	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("../../public/"))))

	// Serve index page on all unhandled routes
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, router *http.Request) {
		http.ServeFile(w, router, "../../public/index.html")
	})

	fmt.Println("http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
