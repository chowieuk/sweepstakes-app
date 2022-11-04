package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamResponse is used to construct requests and parse responses which include an array of teams
type TeamResponse struct {
	Status string     `json:"status"`
	Teams  []TeamData `json:"data"`
}

// TeamData is a reduced model of data associated with a team participating in the 2022 world cup
// It essentially fits the data supplied from the API we're using for world cup data
// https://github.com/raminmr/free-api-worldcup2022

type TeamData struct {
	ObjectID primitive.ObjectID `json:"_id" bson:"_id"`
	Name     string             `json:"name_en" bson:"name_en"`
	Flag     string             `json:"flag" bson:"flag"`
	FifaCode string             `json:"fifa_code" bson:"fifa_code"`
	ISO2     string             `json:"iso2" bson:"iso2"`
	Group    string             `json:"groups" bson:"groups"`
	ID       string             `json:"id" bson:"id"`
	User_id  string             `json:"user_id,omitempty" bson:"user_id"`
	User     []User             `json:"user,omitempty" bson:"user,omitempty"`
}
