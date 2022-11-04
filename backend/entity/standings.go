package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StandingsResponse is used to construct requests and parse responses which include an array of standings
type StandingsResponse struct {
	Status    string          `json:"status"`
	Standings []StandingsData `json:"data"`
}

// StandingsData is a model for the standings data for each group. Mirrors the model used in https://github.com/raminmr/free-api-worldcup2022
type StandingsData struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Group     string             `json:"group" bson:"group"`
	TeamsData []TeamStandingData `json:"teams" bson:"teams"`
}

// TeamStandingData is a model for the standings data for each team. Mirrors the model used in https://github.com/raminmr/free-api-worldcup2022
type TeamStandingData struct {
	Team_id         string `json:"team_id" bson:"team_id"`
	Matches_played  string `json:"mp" bson:"mp"`
	Wins            string `json:"w" bson:"w"`
	Losses          string `json:"l" bson:"l"`
	Points          string `json:"pts" bson:"pts"`
	Goals_for       string `json:"gf" bson:"gf"`
	Goals_against   string `json:"ga" bson:"ga"`
	Goal_difference string `json:"gd" bson:"gd"`
	Name            string `json:"name_en" bson:"name_en"`
	Flag            string `json:"flag" bson:"flag"`
	User            []User `json:"user,omitempty" bson:"user,omitempty"`
}
