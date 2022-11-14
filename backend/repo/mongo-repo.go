package repo

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/chowieuk/sweepstakes-app/backend/entity"

	log "github.com/go-pkgz/lgr"
	"github.com/joho/godotenv"

	"github.com/go-pkgz/auth/token"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// UserInCollection checks if the user id of the token exists in the given mongodb collection
func UserInCollection(userCollection *mongo.Collection, user token.User) (bool, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	log.Printf("[INFO] checking if user %v (id %v) exists in mongodb", user.Name, user.ID)

	count, err := userCollection.CountDocuments(ctx, bson.M{"id": user.ID})
	// log.Printf("[DEBUG] number of records for user %v (id %v) = %v", user.Name, user.ID, count)
	defer cancel()
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// AddSocialUser Adds a social login user (Google or Facebook) to our users collection
func AddSocialUser(userCollection *mongo.Collection, user token.User) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var newUser entity.SocialUser
	newUser.ObjectID = primitive.NewObjectID()
	newUser.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	newUser.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	newUser.Full_Name = user.Name
	newUser.Email = user.Email
	newUser.ID = user.ID

	resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, newUser)
	defer cancel()
	if insertErr != nil {
		return insertErr
	}
	log.Printf("[INFO] successfully added %s to mongodb %s", newUser.Full_Name, resultInsertionNumber)
	return nil
}

// UpdateSocialUserWithTeam updates a social user in our userCollection with their assigned team
func UpdateSocialUserWithTeam(userCollection *mongo.Collection, user token.User, team entity.TeamData) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	log.Printf("[DEBUG] updating social user entry in db with team id")

	result, err := userCollection.UpdateOne(ctx, bson.D{{Key: "id", Value: user.ID}}, bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "team_id",
			Value: team.Team_id}}}})
	defer cancel()
	if !(result.ModifiedCount > 0) {
		log.Printf("[DEBUG] no records were modified")
		log.Printf("[DEBUG] no record added when attempting to update user %s (ID: %s) with team %s (ID: %s) ", user.Email, user.ID, team.Name, team.Team_id)
	}

	if err != nil {
		log.Printf("[DEBUG] failed when attempting to update user %s (ID: %s) with team %s (ID: %s) ", user.Email, user.ID, team.Name, team.Team_id)
		return err
	}
	return nil
}

// RandomUnassignedTeam returns a random unassigned team if one is available, or a team named "Waiting List" and a "randomUnassignedTeam" error if no team is available
func RandomUnassignedTeam(teamCollection *mongo.Collection, ctx context.Context) (entity.TeamData, error) {

	var team entity.TeamData

	// Using Aggregation / samples

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "user_id", Value: nil}}}}
	sampleStage := bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}}

	cursor, err := teamCollection.Aggregate(ctx, mongo.Pipeline{matchStage, sampleStage})

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

// GetUserTeam provides the team associated with a token user
func GetUserTeam(teamCollection *mongo.Collection, user token.User) (entity.TeamData, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var team entity.TeamData
	// attempt to find a team assigned to the user
	err := teamCollection.FindOne(ctx, bson.M{"user_id": user.ID}).Decode(&team)
	defer cancel()
	if err != nil {
		log.Printf("[DEBUG] failed when attempting to find an available team")
		return entity.TeamData{}, err
	}
	return team, nil
}

// CheckTeamAvailability returns true if a team is available
func CheckTeamAvailability(teamCollection *mongo.Collection) (bool, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	log.Printf("[DEBUG] checking amount of available teams")

	count, err := teamCollection.CountDocuments(ctx, bson.M{"user_id": primitive.Null{}})
	defer cancel()
	if err != nil {
		return false, err
	}
	if count > 0 {
		log.Printf("[DEBUG] number of teams available: %v", count)
		return true, nil
	}
	log.Printf("[DEBUG] !!! NO TEAMS AVAILABLE !!!")
	return false, nil
}

// AllocateTeamSocial assigns a team to a Social user
func AllocateTeamSocial(teamCollection *mongo.Collection, user token.User) (entity.TeamData, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var team entity.TeamData

	log.Printf("[INFO] attempting to allocate team to social user")

	// Note: This should only be executed after a check for team availability
	team, err := RandomUnassignedTeam(teamCollection, ctx)
	defer cancel()
	if err != nil {
		log.Printf("[DEBUG] failed when attempting to find an available team")
		return team, err
	}

	result, err := teamCollection.UpdateByID(ctx, team.ObjectID.Hex(), bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "user_id",
			Value: user.ID}}}})
	defer cancel()
	if !(result.ModifiedCount > 0) {
		log.Printf("[DEBUG] no team documents were modified")
	}

	if err != nil {
		log.Printf("[DEBUG] failed when attempting to update team %s (Object ID: %s) with user %s id %s ", team.Name, team.Team_id, user.Name, user.ID)
		return entity.TeamData{}, err
	}

	user.SetStrAttr("team_name", team.Name)
	user.SetStrAttr("team_flag", team.Flag)
	log.Printf("[INFO] successfully allocated %s (id %s) to %s (id %s)\n", team.Name, team.Team_id, user.Name, user.ID)
	// log.Printf("[INFO] available teams left: %s)\n",)
	return team, nil
}
