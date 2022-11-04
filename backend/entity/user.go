package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is a reduced model for new registrants that will be inserted into the DB
type User struct {
	ObjectID   *primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Created_at *time.Time          `json:"created_at,omitempty"`
	Updated_at *time.Time          `json:"updated_at,omitempty"`
	Full_Name  string              `json:"full_name"`
	Email      string              `json:"email"`
	Password   *string             `json:"password,omitempty"`
	ID         string              `json:"id,omitempty"`
	Team_id    string              `json:"team_id"`
}

// SocialUser is a reduced model for new registrants that will be inserted into the DB
type SocialUser struct {
	ObjectID   primitive.ObjectID `bson:"_id"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Full_Name  string             `json:"full_name"`
	Email      string             `json:"email"`
	ID         string             `json:"id"`
	Team_id    string             `json:"team_id"`
}
