package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MatchResponse is used to construct requests and parse responses which include an array of matches
type MatchResponse struct {
	Error   string      `json:"error,omitempty"`
	Status  string      `json:"status"`
	Matches []MatchData `json:"data"`
}

// MatchData is a reduced model of data associated with matches occuring in the 2022 world cup
// It essentially fits the data supplied from the API we're using for world cup data
// https://github.com/raminmr/free-api-worldcup2022
type MatchData struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	Away_score     int                `json:"away_score" bson:"away_score"`
	Away_scorers   []string           `json:"away_scorers" bson:"away_scorers"`
	Away_team_id   string             `json:"away_team_id" bson:"away_team_id"`
	Finished       string             `json:"finished" bson:"finished"`
	Group          string             `json:"group" bson:"group"`
	Home_score     int                `json:"home_score" bson:"home_score"`
	Home_scorers   []string           `json:"home_scorers" bson:"home_scorers"`
	Home_team_id   string             `json:"home_team_id" bson:"home_team_id"`
	Match_Id       string             `json:"id" bson:"id"`
	Local_date     string             `json:"local_date" bson:"local_date"` // TODO: investigate date normalization
	Matchday       string             `json:"matchday" bson:"matchday"`
	Stadium_id     string             `json:"stadium_id" bson:"stadium_id"`
	Time_elapsed   string             `json:"time_elapsed" bson:"time_elapsed"`
	Type           string             `json:"type" bson:"type"`
	Home_team_name string             `json:"home_team_en" bson:"home_team_en"`
	Away_team_name string             `json:"away_team_en" bson:"away_team_en"`
	Home_flag      string             `json:"home_flag" bson:"home_flag"`
	Away_flag      string             `json:"away_flag" bson:"away_flag"`
	Away_User      []User             `json:"away_user,omitempty" bson:"away_user,omitempty"`
	Home_User      []User             `json:"home_user,omitempty" bson:"home_user,omitempty"`
}
