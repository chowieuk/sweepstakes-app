package repo

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/chowieuk/sweepstakes-app/backend/entity"

	log "github.com/go-pkgz/lgr"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBinstance func creates a new mongo client with URI provided in .env
func DBinstance() *mongo.Client {
	err := godotenv.Load(".env.production.local")

	if err != nil {
		log.Fatalf("[ERROR] Error loading .env file")
	}

	MongoDb := os.Getenv("MONGO_DB_URL")
	//fmt.Fprintf("mongodb+srv://%s:%s@cluster0.1gyiabr.mongodb.net/?retryWrites=true&w=majority",os.Getenv("MONGO_DB_USER"),os.Getenv("MONGO_DB_PASSWORD"))
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Fatalf("[ERROR] Error creating mongo client", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("[ERROR] Error connecting to mongo client", err)
	}
	log.Printf("[INFO] Connected to MongoDB!")

	return client
}

// OpenCollection is a function that makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database(os.Getenv("MONGO_DB_NAME")).Collection(collectionName)

	return collection
}

func RandomUnassignedTeam(collection *mongo.Collection, ctx context.Context) (entity.TeamData, error) {

	var team entity.TeamData

	// Using Aggregation / samples

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "user_id", Value: nil}}}}
	sampleStage := bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, sampleStage})

	if cursor.RemainingBatchLength() < 1 {
		log.Printf("[DEBUG] no more teams are available")
		return entity.TeamData{Name: "Waiting List"}, errors.New("randomUnassignedTeam: all teams are assigned")
	}
	if err != nil {
		log.Printf("[DEBUG] error durring aggregation: %v", err)
		return entity.TeamData{}, err
	}

	for cursor.Next(ctx) {
		if err := cursor.Decode(&team); err != nil {
			log.Printf("[DEBUG] error during decode: %v", err)
			return entity.TeamData{}, err
		}
	}
	return team, nil
}
