package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chowieuk/sweepstakes-app/backend/entity"
	"github.com/chowieuk/sweepstakes-app/backend/repo"
	"github.com/chowieuk/sweepstakes-app/backend/service"

	"github.com/robfig/cron/v3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/go-pkgz/auth/middleware"
	"github.com/go-pkgz/auth/token"
	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/rest"
	"github.com/go-pkgz/rest/logger"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Client Database instance
var Client *mongo.Client = repo.DBinstance()

var userCollection *mongo.Collection = repo.OpenCollection(Client, "users")
var teamCollection *mongo.Collection = repo.OpenCollection(Client, "teams")
var matchesCollection *mongo.Collection = repo.OpenCollection(Client, "matches")
var standingsCollection *mongo.Collection = repo.OpenCollection(Client, "standings")

var apiToken, apiUser, apiPass string

func main() {

	err := godotenv.Load(".env.production.local")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	if apiUser = os.Getenv("FIFA_API_USER"); apiUser == "" {
		log.Fatalf("[FATAL] You must set your 'FIFA_API_USER' environmental variable.")
	}
	if apiPass = os.Getenv("FIFA_API_PASS"); apiPass == "" {
		log.Fatalf("[FATAL] You must set your 'FIFA_API_PASS' environmental variable.")
	}

	if apiToken = os.Getenv("FIFA_API_TOKEN"); apiToken == "" {
		log.Print("[INFO] You haven't set your 'FIFA_API_TOKEN' environmental variable.")
		var err error
		apiToken, err = repo.LoginToAPI(apiUser, apiPass)
		if err != nil {
			log.Fatalf("[FATAL] Couldn't acquire API token: %v", err)
		}
	}

	log.Setup(log.Debug, log.Msec, log.LevelBraces, log.CallerFile, log.CallerFunc) // setup default logger with go-pkgz/lgr

	// initialize cron job for data update
	c := cron.New()
	// matchesInterval runs every 30th minute from 10am to 7pm London time
	matchesInterval := "CRON_TZ=Europe/London */30 10-19 * * *"
	// standingsInterval runs every 32nd minute from 10am to 7pm London time
	standingsInterval := "CRON_TZ=Europe/London */32 10-19 * * *"
	log.Printf("[INFO]\nCron job set to update data on the following schedule:\nMatches: %v\nStandings data '%v'", matchesInterval, standingsInterval)

	matchesCronJob := cron.FuncJob(func() {
		log.Printf("[INFO] Initiating scheduled matches data update")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		matches, err := repo.FetchMatches(apiToken, apiUser, apiPass)
		if err != nil {
			log.Printf("[CRITICAL] error fetching match data:", err)
			return
		}

		if err := repo.UpdateMatches(matchesCollection, matches, ctx); err != nil {
			log.Fatalf("[CRITICAL] recieved error updating match data: ", err)
			return
		}
	})

	standingsCronJob := cron.FuncJob(func() {
		log.Printf("[INFO] Initiating scheduled standings data update")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		standings, err := repo.FetchStandings(apiToken, apiUser, apiPass)
		if err != nil {
			log.Printf("[CRITICAL] error fetching standings data:", err)
			return
		}

		if err := repo.UpdateStandings(standingsCollection, standings, ctx); err != nil {
			log.Fatalf("[CRITICAL] recieved error updating standings data: ", err)
			return
		}
	})

	c.AddJob(matchesInterval, matchesCronJob)
	c.AddJob(standingsInterval, standingsCronJob)
	c.Start()

	authService := service.InitializeAuth(userCollection, teamCollection)

	// retrieve auth middleware
	m := authService.Middleware()

	// setup http server
	router := chi.NewRouter()
	// add some external middlewares from go-pkgz/rest
	router.Use(rest.AppInfo("sweepstakes", "chowieuk, patrickreynoldscoding", "1.0.0"), rest.Ping)
	router.Use(logger.New(logger.Log(log.Default()), logger.WithBody, logger.Prefix("[INFO]")).Handler) // log all http requests
	// Basic CORS
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://chowie.uk", "http://localhost:3000", "http://localhost:8080"}, // Use this to allow specific origin hosts
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		Debug:            false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	router.Post("/register", registrationHandler) // post registration route
	router.Group(func(r chi.Router) {
		r.Use(rest.Throttle(100))
		router.Get("/api/v1/availableteams", availableTeamCount) // get available teams route
	})
	router.Group(func(r chi.Router) {
		r.Use(m.Auth)
		r.Use(m.UpdateUser(middleware.UserUpdFunc(func(user token.User) token.User {
			return user
		})))
		r.Get("/private_data", protectedDataHandler)                            // protected api
		r.Get("/api/v1/team", allTeamsResponseHandler)                          // data for all teams
		r.Get("/api/v1/team/{id}", singleTeamResponseHandler)                   // data for a specific team by team id
		r.Get("/api/v1/match", allMatchesResponseHandler)                       // data for all matches
		r.Get("/api/v1/match/{id}", singleMatchResponseHandler)                 // data for a single match by match id
		r.Get("/api/v1/match/day/{day}", byDayMatchResponseHandler)             // data for all matches on a given day
		r.Post("/api/v1/match/date", byDateMatchResponseHandler)                // data for all matches on a given date. Post request should include a body in the form {"date":"mm/dd/yyyy"}
		r.Get("/api/v1/standings", allStandingsResponseHandler)                 // data for all standings
		r.Get("/api/v1/standings/group/{group}", groupStandingsResponseHandler) // data for standings of a specific group
		r.Get("/api/v1/standings/team/{id}", teamStandingsResponseHandler)      // data for standings of a specific team
	})

	// declare custom 404
	// custom404, err := os.Open(filepath.Join(workDir, "build/custom404.html"))
	// FsOptCustom404(custom404)

	// serve static build files under /
	FS, err := rest.NewFileServer("/", "build", rest.FsOptSPA)

	if err != nil {
		log.Printf("Error initializaing rest.NewFileServer: ", err)
	}

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		FS.ServeHTTP(w, r)
	})

	// serve static auth example front end files under /web
	workDir, _ := os.Getwd()
	authFilesDir := filepath.Join(workDir, "auth-example-frontend")
	fileServer(router, "/web", http.Dir(authFilesDir))

	// setup auth routes
	authRoutes, avaRoutes := authService.Handlers()
	router.Mount("/auth", authRoutes)  // add auth handlers
	router.Mount("/avatar", avaRoutes) // add avatar handler

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Printf("[PANIC] failed to start http server, %v", err)
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve static files from a http.FileSystem.
// Borrowed from https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	log.Printf("[INFO] serving static files from %v", root)
	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

// attempts to register with given details, and returns an error / success

func registrationHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("[INFO] client attempting registration")

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var user entity.User
	err := json.NewDecoder(r.Body).Decode(&user)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusBadRequest, err, "failed to parse registration")
		return
	}

	log.Printf("[INFO] checking if provided email exists in mongodb")

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to lookup username")
		return
	}

	if count > 0 {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusForbidden, nil, "the username provided already exists")
		return
	}

	log.Printf("[INFO] populating client entity and allocating team")

	password := service.HashPassword(*user.Password)
	user.Password = &password
	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Created_at = &now
	user.Updated_at = &now
	newId := primitive.NewObjectID()
	user.ObjectID = &newId
	user.ID = "mongo_" + token.HashID(sha1.New(), user.Email)

	team, err := repo.RandomUnassignedTeam(teamCollection, ctx)

	if err != nil && err != repo.ErrNoTeams {
		log.Printf("[DEBUG] error getting random team : %v", err)
		return
	}

	if err == repo.ErrNoTeams {
		//TODO: consider waiting list user story?
		log.Printf("[DEBUG] all teams allocated. Adding user to waiting list: %v", err)
		user.Team_id = ""
	} else {
		err = allocateTeam(team, &user, ctx)
		if err != nil {
			rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to allocate team")
			return
		}
	}

	resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
	defer cancel()
	if insertErr != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, insertErr, "failed to create new user")
		return
	}
	log.Printf("[INFO] successfully added %s to mongodb %s", user.Email, resultInsertionNumber)
	rest.RenderJSON(w, rest.JSON{"message": "User successfully created"})
	// Alternatively we could respond with our user:
	// rest.RenderJSON(w, &user)
}

func allocateTeam(team entity.TeamData, user *entity.User, ctx context.Context) error {

	result, err := teamCollection.UpdateByID(ctx, team.ObjectID.Hex(), bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "user_id",
			Value: user.ID}}}})

	if !(result.ModifiedCount > 0) {
		log.Printf("[DEBUG] no records were modified")
	}

	if err != nil {
		log.Printf("[DEBUG] failed when attempting to update team %s (Object ID: %s) with user id %s ", team.Name, team.Team_id, user.ID)
		return err
	}

	user.Team_id = team.Team_id
	log.Printf("[INFO] successfully allocated %s (id %s) to %s (id %s)\n", team.Name, team.Team_id, user.Email, user.ID)
	return nil
}

func availableTeamCount(w http.ResponseWriter, r *http.Request) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	count, err := teamCollection.CountDocuments(ctx, bson.M{"user_id": primitive.Null{}})
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch available teams")
		return
	}

	// _, _ = w.Write([]byte(strconv.Itoa(int(count))))

	res := struct {
		TS             time.Time `json:"ts"`
		AvailableTeams int64     `json:"availableTeams"`
	}{
		TS:             time.Now(),
		AvailableTeams: count,
	}

	rest.RenderJSON(w, res)
}

// GET /private_data returns json with user info and ts
func protectedDataHandler(w http.ResponseWriter, r *http.Request) {

	userInfo, err := token.GetUserInfo(r)
	if err != nil {
		log.Printf("failed to get user info, %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res := struct {
		TS     time.Time  `json:"ts"`
		Field1 string     `json:"fld1"`
		Field2 int        `json:"fld2"`
		User   token.User `json:"userInfo"`
	}{
		TS:     time.Now(),
		Field1: "some private thing",
		Field2: 42,
		User:   userInfo,
	}

	rest.RenderJSON(w, res)
}

// singleTeamResponseHandler responds with JSON for a single team by team_id, with a lookup for the associated user
func singleTeamResponseHandler(w http.ResponseWriter, r *http.Request) {

	team_id := chi.URLParam(r, "id")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "id", Value: team_id}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "user_id"},
			{Key: "foreignField", Value: "id"},
			{Key: "as", Value: "user"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "user_id", Value: 0},
			{Key: "user.password", Value: 0},
			{Key: "user._id", Value: 0},
			{Key: "user.id", Value: 0},
			{Key: "user.team_id", Value: 0},
			{Key: "user.created_at", Value: 0},
			{Key: "user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	cursor, err := teamCollection.Aggregate(ctx, pipeline)
	defer cancel()

	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch team")
		return
	}

	if cursor.RemainingBatchLength() < 1 {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusBadRequest, err, "no results for given team id")
		return
	}

	cursor.RemainingBatchLength()

	var res entity.TeamResponse

	if err = cursor.All(ctx, &res.Teams); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse team")
		return
	}
	defer cancel()
	if res.Teams != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// allTeamsResponseHandler responds with JSON for all teams, with a lookup for the user associated with that team
func allTeamsResponseHandler(w http.ResponseWriter, r *http.Request) {

	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "user_id"},
			{Key: "foreignField", Value: "id"},
			{Key: "as", Value: "user"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "user_id", Value: 0},
			{Key: "user.password", Value: 0},
			{Key: "user._id", Value: 0},
			{Key: "user.id", Value: 0},
			{Key: "user.team_id", Value: 0},
			{Key: "user.created_at", Value: 0},
			{Key: "user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := teamCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	var res entity.TeamResponse

	if err = cursor.All(ctx, &res.Teams); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Teams != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// singleMatchResponseHandler responds with JSON for a single team by match_id, with a lookup for the associated user
func singleMatchResponseHandler(w http.ResponseWriter, r *http.Request) {

	match_id := chi.URLParam(r, "id")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "id", Value: match_id}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "away_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "away_user"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "home_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "home_user"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "home_user.password", Value: 0},
			{Key: "home_user._id", Value: 0},
			{Key: "home_user.id", Value: 0},
			{Key: "home_user.team_id", Value: 0},
			{Key: "home_user.created_at", Value: 0},
			{Key: "home_user.updated_at", Value: 0},
			{Key: "away_user.password", Value: 0},
			{Key: "away_user._id", Value: 0},
			{Key: "away_user.id", Value: 0},
			{Key: "away_user.team_id", Value: 0},
			{Key: "away_user.created_at", Value: 0},
			{Key: "away_user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := matchesCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	var res entity.MatchResponse

	if err = cursor.All(ctx, &res.Matches); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Matches != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// byDayMatchResponseHandler responds with JSON for a all matchs on a specific matchday, with a lookup for the associated user
func byDayMatchResponseHandler(w http.ResponseWriter, r *http.Request) {

	matchday := chi.URLParam(r, "day")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "matchday", Value: matchday}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "away_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "away_user"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "home_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "home_user"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "home_user.password", Value: 0},
			{Key: "home_user._id", Value: 0},
			{Key: "home_user.id", Value: 0},
			{Key: "home_user.team_id", Value: 0},
			{Key: "home_user.created_at", Value: 0},
			{Key: "home_user.updated_at", Value: 0},
			{Key: "away_user.password", Value: 0},
			{Key: "away_user._id", Value: 0},
			{Key: "away_user.id", Value: 0},
			{Key: "away_user.team_id", Value: 0},
			{Key: "away_user.created_at", Value: 0},
			{Key: "away_user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := matchesCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	var res entity.MatchResponse

	if err = cursor.All(ctx, &res.Matches); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Matches != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// byDateMatchResponseHandler responds with JSON of the matches occuring on the date specified the POST request body, with a lookup for the associated user. Date should be in the form {"date":"mm/dd/yyyy"}.
func byDateMatchResponseHandler(w http.ResponseWriter, r *http.Request) {

	request := struct {
		Date string `json:"date"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed decoding date")
		return
	}

	pipeline := mongo.Pipeline{
		{{Key: "$addFields", Value: bson.D{
			{Key: "uk_date", Value: bson.D{
				{Key: "$dateToString", Value: bson.D{
					{Key: "format", Value: "%m/%d/%Y"},
					{Key: "timezone", Value: "Europe/London"},
					{Key: "date", Value: bson.D{
						{Key: "$dateFromString", Value: bson.D{
							{Key: "dateString", Value: "$local_date"},
							{Key: "format", Value: "%m/%d/%Y %H:%M"},
							{Key: "timezone", Value: "Asia/Qatar"},
							{Key: "onError", Value: "cannot parse date from API"},
							{Key: "onNull", Value: "null value in data from API"},
						}}}}}}}}}}},
		{{Key: "$match", Value: bson.D{{Key: "uk_date", Value: request.Date}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "away_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "away_user"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "home_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "home_user"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "home_user.password", Value: 0},
			{Key: "home_user._id", Value: 0},
			{Key: "home_user.id", Value: 0},
			{Key: "home_user.team_id", Value: 0},
			{Key: "home_user.created_at", Value: 0},
			{Key: "home_user.updated_at", Value: 0},
			{Key: "away_user.password", Value: 0},
			{Key: "away_user._id", Value: 0},
			{Key: "away_user.id", Value: 0},
			{Key: "away_user.team_id", Value: 0},
			{Key: "away_user.created_at", Value: 0},
			{Key: "away_user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := matchesCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	if cursor.RemainingBatchLength() < 1 {
		log.Printf(request.Date)
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch matches. maybe check your date")
	}

	var res entity.MatchResponse

	if err = cursor.All(ctx, &res.Matches); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Matches != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// allMatchesResponseHandler responds with JSON of all matches in the matches collection, with a lookup for the users associated with the home & away teams
func allMatchesResponseHandler(w http.ResponseWriter, r *http.Request) {

	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "away_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "away_user"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "home_team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "home_user"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "home_user.password", Value: 0},
			{Key: "home_user._id", Value: 0},
			{Key: "home_user.id", Value: 0},
			{Key: "home_user.team_id", Value: 0},
			{Key: "home_user.created_at", Value: 0},
			{Key: "home_user.updated_at", Value: 0},
			{Key: "away_user.password", Value: 0},
			{Key: "away_user._id", Value: 0},
			{Key: "away_user.id", Value: 0},
			{Key: "away_user.team_id", Value: 0},
			{Key: "away_user.created_at", Value: 0},
			{Key: "away_user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := matchesCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	var res entity.MatchResponse

	if err = cursor.All(ctx, &res.Matches); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Matches != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// allStandingsResponseHandler responds with JSON of all standings in the standings collection, with a lookup to include the user associated with each team. The results of this lookup have been mapped onto an an additionl field combining objects in the teams array with user objects. This additional field replaces the original teams field
// seems like a workaround, but is necessary if we want to conform to the schema of the source API (https://jira.mongodb.org/browse/SERVER-42306?focusedCommentId=2348528&page=com.atlassian.jira.plugin.system.issuetabpanels:comment-tabpanel#comment-2348528)
func allStandingsResponseHandler(w http.ResponseWriter, r *http.Request) {
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "teams.team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "users"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "teams", Value: bson.D{
				{Key: "$map", Value: bson.D{
					{Key: "input", Value: "$teams"},
					{Key: "as", Value: "teams"},
					{Key: "in", Value: bson.D{
						{Key: "$mergeObjects", Value: bson.A{"$$teams", bson.D{
							{Key: "user", Value: bson.D{
								{Key: "$filter", Value: bson.D{
									{Key: "input", Value: "$users"},
									{Key: "cond", Value: bson.D{
										{Key: "$eq", Value: bson.A{
											"$$teams.team_id",
											"$$this.team_id",
										},
										}}}}}}}}}}}}}}}}}}},
		{{Key: "$project", Value: bson.D{
			{Key: "users", Value: 0},
			{Key: "teams.user_id", Value: 0},
			{Key: "teams.user.team_id", Value: 0},
			{Key: "teams.user.password", Value: 0},
			{Key: "teams.user._id", Value: 0},
			{Key: "teams.user.id", Value: 0},
			{Key: "teams.user.created_at", Value: 0},
			{Key: "teams.user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := standingsCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	var res entity.StandingsResponse

	if err = cursor.All(ctx, &res.Standings); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Standings != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// allStandingsResponseHandler responds with JSON of standings by group, with a lookup to include the user associated with each team. The results of this lookup have been mapped onto an an additionl field combining objects in the teams array with user objects. This additional field replaces the original teams field
func groupStandingsResponseHandler(w http.ResponseWriter, r *http.Request) {

	group := chi.URLParam(r, "group")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "group", Value: group}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "teams.team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "users"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "teams", Value: bson.D{
				{Key: "$map", Value: bson.D{
					{Key: "input", Value: "$teams"},
					{Key: "as", Value: "teams"},
					{Key: "in", Value: bson.D{
						{Key: "$mergeObjects", Value: bson.A{"$$teams", bson.D{
							{Key: "user", Value: bson.D{
								{Key: "$filter", Value: bson.D{
									{Key: "input", Value: "$users"},
									{Key: "cond", Value: bson.D{
										{Key: "$eq", Value: bson.A{
											"$$teams.team_id",
											"$$this.team_id",
										},
										}}}}}}}}}}}}}}}}}}},
		{{Key: "$project", Value: bson.D{
			{Key: "users", Value: 0},
			{Key: "teams.user_id", Value: 0},
			{Key: "teams.user.team_id", Value: 0},
			{Key: "teams.user.password", Value: 0},
			{Key: "teams.user._id", Value: 0},
			{Key: "teams.user.id", Value: 0},
			{Key: "teams.user.created_at", Value: 0},
			{Key: "teams.user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := standingsCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	if cursor.RemainingBatchLength() < 1 {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch standings, check your group")
		return
	}

	var res entity.StandingsResponse

	if err = cursor.All(ctx, &res.Standings); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Standings != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}

// allStandingsResponseHandler responds with JSON of standings by team_id, with a lookup to include the user associated with each team. The results of this lookup have been mapped onto an an additionl field combining objects in the teams array with user objects. This additional field replaces the original teams field
func teamStandingsResponseHandler(w http.ResponseWriter, r *http.Request) {

	team_id := chi.URLParam(r, "id")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "teams.team_id", Value: team_id}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "teams.team_id"},
			{Key: "foreignField", Value: "team_id"},
			{Key: "as", Value: "users"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "teams", Value: bson.D{
				{Key: "$map", Value: bson.D{
					{Key: "input", Value: "$teams"},
					{Key: "as", Value: "teams"},
					{Key: "in", Value: bson.D{
						{Key: "$mergeObjects", Value: bson.A{"$$teams", bson.D{
							{Key: "user", Value: bson.D{
								{Key: "$filter", Value: bson.D{
									{Key: "input", Value: "$users"},
									{Key: "cond", Value: bson.D{
										{Key: "$eq", Value: bson.A{
											"$$teams.team_id",
											"$$this.team_id",
										},
										}}}}}}}}}}}}}}}}}}},
		{{Key: "$project", Value: bson.D{
			{Key: "users", Value: 0},
			{Key: "teams.user_id", Value: 0},
			{Key: "teams.user.team_id", Value: 0},
			{Key: "teams.user.password", Value: 0},
			{Key: "teams.user._id", Value: 0},
			{Key: "teams.user.id", Value: 0},
			{Key: "teams.user.created_at", Value: 0},
			{Key: "teams.user.updated_at", Value: 0},
		}}}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := standingsCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
		return
	}

	var res entity.StandingsResponse

	if err = cursor.All(ctx, &res.Standings); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
		return
	}
	defer cancel()
	if res.Standings != nil {
		res.Status = "success"
		rest.RenderJSON(w, res)
	}
}
