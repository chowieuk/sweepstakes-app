package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chowieuk/sweepstakes-app/backend/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TokenResponse struct {
	Error  string `json:"error"`
	Status string `json:"status"`
	Data   Token  `json:"data"`
}

type Token struct {
	Token string `json:"token"`
}

func LoginToAPI(user, pass string) (token string, err error) {

	log.Println("[INFO] Attempting to refresh API token")

	endpoint := "http://api.cup2022.ir/api/v1/user/login"
	// curl --location --request POST 'http://api.cup2022.ir/api/v1/user/login' \
	// --header 'Content-Type: application/json' \
	// --data-raw '{
	// "email": "email",
	// "password": "pass"
	// }'

	worldcupClient := http.Client{
		Timeout: time.Second * 5, // Timeout after 5 seconds
	}

	loginData := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    user,
		Password: pass,
	}

	loginDataJSON, err := json.Marshal(loginData)
	if err != nil {
		log.Printf("error while marshalling request body: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(loginDataJSON))
	if err != nil {
		log.Printf("error while creating new request: %v", err)
		return
	}

	req.Header.Set("User-Agent", "sweepstakes-app")
	req.Header.Set("Content-Type", "application/json")

	res, err := worldcupClient.Do(req)
	if err != nil {
		log.Printf("error while submitting POST request: %v", err)
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	var tokenResponse *TokenResponse
	err = json.NewDecoder(res.Body).Decode(&tokenResponse)
	if err != nil {
		log.Printf("could not unmarshal json: %s\n", err)
		return
	}

	if tokenResponse.Error != "" {
		log.Printf("World Cup API responded with error: %v", tokenResponse.Error)
		err = errors.New(tokenResponse.Error)
		return
	}

	log.Printf("World Cup API responded with status: %#v\n", tokenResponse.Status)
	log.Printf("Token in response: %#v\n", tokenResponse.Data.Token)
	return tokenResponse.Data.Token, nil
}

func FetchMatches(apiToken, apiUser, apiPass string) (matches []entity.MatchData, err error) {

	endpoint := "http://api.cup2022.ir/api/v1/match"
	// curl --location --request GET 'http://api.cup2022.ir/api/v1/match' --header 'Authorization: Bearer '$WORLD_CUP_TOKEN'' --header 'Content-Type: application/json'

	worldcupClient := http.Client{
		Timeout: time.Second * 5, // Timeout after 5 seconds
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("[MATCHES] error while creating new request: %v", err)
		return
	}

	req.Header.Set("User-Agent", "sweepstakes-app")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	res, err := worldcupClient.Do(req)
	if err != nil {
		log.Printf("[MATCHES] error while submitting GET request: %v", err)
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	var matchResponse *entity.MatchResponse
	err = json.NewDecoder(res.Body).Decode(&matchResponse)
	if err != nil {
		log.Printf("[MATCHES] could not unmarshal json: %s\n", err)
		return
	}

	if matchResponse.Error != "" {
		if matchResponse.Error == "Not authorized to access this resource" {
			// Try refreshing the token
			apiToken, err = LoginToAPI(apiUser, apiPass)
			if err != nil {
				log.Printf("[MATCHES] error while acquiring new API token: %v", err)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

			res, err = worldcupClient.Do(req)
			if err != nil {
				log.Printf("[MATCHES] error while submitting GET request: %v", err)
				return
			}

			if res.Body != nil {
				defer res.Body.Close()
			}

			err = json.NewDecoder(res.Body).Decode(&matchResponse)
			if err != nil {
				log.Printf("[MATCHES] could not unmarshal json: %s\n", err)
				return
			}
			log.Printf("[MATCHES] World Cup API responded with status: %#v\n", matchResponse.Status)
			log.Printf("[MATCHES] records in matches response: %#v\n", len(matchResponse.Matches))

			return matchResponse.Matches, nil
		}
		log.Printf("[MATCHES] World Cup API responded with unknown error: %v", matchResponse.Error)
	}

	log.Printf("[MATCHES] World Cup API responded with status: %#v\n", matchResponse.Status)
	log.Printf("[MATCHES] records in matches response: %#v\n", len(matchResponse.Matches))

	return matchResponse.Matches, nil
}

func UpdateMatches(matchesCollection *mongo.Collection, matches []entity.MatchData, ctx context.Context) (err error) {

	log.Printf("[MATCHES] attempting to update / insert %v records provided in response\n", len(matches))

	var updatedCount int64

	// Will insert a document if there are no matches for our query filter
	opts := options.Update().SetUpsert(true)

	for _, match := range matches {
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "away_score", Value: match.Away_score},
			{Key: "away_scorers", Value: match.Away_scorers},
			{Key: "away_team_id", Value: match.Away_team_id},
			{Key: "finished", Value: match.Finished},
			{Key: "group", Value: match.Group},
			{Key: "home_score", Value: match.Home_score},
			{Key: "home_scorers", Value: match.Home_scorers},
			{Key: "home_team_id", Value: match.Home_team_id},
			{Key: "id", Value: match.Match_Id},
			{Key: "local_date", Value: match.Local_date},
			{Key: "matchday", Value: match.Matchday},
			{Key: "stadium_id", Value: match.Stadium_id},
			{Key: "time_elapsed", Value: match.Time_elapsed},
			{Key: "type", Value: match.Type},
			{Key: "home_team_en", Value: match.Home_team_name},
			{Key: "away_team_en", Value: match.Away_team_name},
			{Key: "home_flag", Value: match.Home_flag},
			{Key: "away_flag", Value: match.Away_flag},
		}}}

		result, err := matchesCollection.UpdateByID(ctx, match.ID.Hex(), update, opts)
		if err != nil {
			log.Printf("[MATCHES] error whilst updating item _id: %v", match.ID.Hex())
		}
		if result.ModifiedCount > 0 {
			log.Printf("[MATCHES] ObjectID(%v) Match %v (%v vs %v) updated.\n", match.ID.Hex(), match.Match_Id, match.Home_team_name, match.Away_team_name)
		}
		updatedCount += result.ModifiedCount
	}
	log.Printf("[MATCHES] number of documents updated: %v\n", updatedCount)
	return
}

func FetchStandings(apiToken, apiUser, apiPass string) (standings []entity.StandingsData, err error) {

	endpoint := "http://api.cup2022.ir/api/v1/standings"
	// curl --location --request GET 'http://api.cup2022.ir/api/v1/standings' --header 'Authorization: Bearer '$WORLD_CUP_TOKEN'' --header 'Content-Type: application/json'

	worldcupClient := http.Client{
		Timeout: time.Second * 5, // Timeout after 5 seconds
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("[STANDINGS] error while creating new request: %v", err)
		return
	}

	req.Header.Set("User-Agent", "sweepstakes-app")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	res, err := worldcupClient.Do(req)
	if err != nil {
		log.Printf("[STANDINGS] error while submitting GET request: %v", err)
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	var standingsResponse *entity.StandingsResponse
	err = json.NewDecoder(res.Body).Decode(&standingsResponse)
	if err != nil {
		log.Printf("[STANDINGS] could not unmarshal json: %s\n", err)
		return
	}

	if standingsResponse.Error != "" {
		if standingsResponse.Error == "Not authorized to access this resource" {
			// Try refreshing the token
			apiToken, err = LoginToAPI(apiUser, apiPass)
			if err != nil {
				log.Printf("[STANDINGS] error while acquiring new API token: %v", err)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

			res, err = worldcupClient.Do(req)
			if err != nil {
				log.Printf("[STANDINGS] error while submitting GET request: %v", err)
				return
			}

			if res.Body != nil {
				defer res.Body.Close()
			}

			err = json.NewDecoder(res.Body).Decode(&standingsResponse)
			if err != nil {
				log.Printf("[STANDINGS] could not unmarshal json: %s\n", err)
				return
			}
			log.Printf("[STANDINGS] World Cup API responded with status: %#v\n", standingsResponse.Status)
			log.Printf("[STANDINGS] records in standings response: %#v\n", len(standingsResponse.Standings))

			return standingsResponse.Standings, nil
		}
		log.Printf("[STANDINGS] World Cup API responded with unknown error: %v", standingsResponse.Error)
	}

	log.Printf("[STANDINGS] World Cup API responded with status: %#v\n", standingsResponse.Status)
	log.Printf("[STANDINGS] records in response: %#v\n", len(standingsResponse.Standings))

	return standingsResponse.Standings, nil
}

func UpdateStandings(standingsCollection *mongo.Collection, standings []entity.StandingsData, ctx context.Context) (err error) {

	log.Printf("[STANDINGS] attempting to update / insert %v records provided in response\n", len(standings))

	var updatedCount int64

	// Will insert a document if there are no standings for our query filter
	opts := options.Update().SetUpsert(true)

	for _, group := range standings {
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "teams", Value: group.TeamsData},
		}}}

		result, err := standingsCollection.UpdateByID(ctx, group.ID.Hex(), update, opts)
		if err != nil {
			log.Printf("[STANDINGS] error whilst updating item ObjectID: %v", group.ID.Hex())
		}

		if result.ModifiedCount > 0 {
			for _, team := range group.TeamsData {
				log.Printf("[STANDINGS] ObjectID(%v) standing entry for %v updated.\n", group.ID.Hex(), team.Name)
			}
		}
		updatedCount += result.ModifiedCount
	}
	log.Printf("[STANDINGS] number of documents updated: %v\n", updatedCount)
	return
}
