package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is a reduced model of objects that will be retrived or inserted into the DB
type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Username   string             `json:"Username"`
	Password   string             `json:"Password"`
	User_id    string             `json:"user_id"`
	//Email      string             `json:"email"`
	//Nation     string             `json:"nation"`
}
