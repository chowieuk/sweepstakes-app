package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is a reduced model for user objects that will be retrived or inserted into our repo
type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Full_Name  string             `json:"full_name"`
	Email      string             `json:"email"`
	Password   string             `json:"password"`
	User_id    string             `json:"user_id"`
	Team_id    string             `json:"team_id"`
}
