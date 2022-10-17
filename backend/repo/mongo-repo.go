package repo

import (
	"context"
	"os"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/joho/godotenv"

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

// OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database(os.Getenv("MONGO_DB_NAME")).Collection(collectionName)

	return collection
}
