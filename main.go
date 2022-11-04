package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chowieuk/sweepstakes-app/backend/entity"
	"github.com/chowieuk/sweepstakes-app/backend/repo"
	"github.com/chowieuk/sweepstakes-app/backend/service"

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

func main() {

	err := godotenv.Load(".env.production.local")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	log.Setup(log.Debug, log.Msec, log.LevelBraces, log.CallerFile, log.CallerFunc) // setup default logger with go-pkgz/lgr

	authService := service.InitializeAuth(userCollection)

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
		r.Use(m.Auth)
		r.Use(m.UpdateUser(middleware.UserUpdFunc(func(user token.User) token.User {
			return user
		})))
		r.Get("/private_data", protectedDataHandler) // protected api
		r.Get("/api/v1/team/{id}", singleTeamResponseHandler)
		r.Get("/api/v1/team", allTeamsResponseHandler)
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

	if err != nil {
		if err == errors.New("randomUnassignedTeam: all teams are assigned") {
			//TODO: consider waiting list user story?
			rest.SendErrorJSON(w, r, log.Default(), http.StatusForbidden, err, "all teams have been allocated")
			return
		}
		log.Printf("[DEBUG] error getting random team : %v", err)
		return
	}

	err = allocateTeam(team, &user, ctx)
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to allocate team")
		return
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
		log.Printf("[DEBUG] failed when attempting to update team %s (Object ID: %s) with user id %s ", team.Name, team.ID, user.ID)
		return err
	}

	user.Team_id = team.ID
	log.Printf("[INFO] successfully allocated %s (id %s) to %s (id %s)\n", team.Name, team.ID, user.Email, user.ID)
	return nil
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

// A request on Team endpoint returns all information on a Team by id

//     Http Method : GET http://chowie.uk/api/v1/team/{id}

func singleTeamResponseHandler(w http.ResponseWriter, r *http.Request) {

	team_id := chi.URLParam(r, "id")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "id", Value: team_id}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "user_id"},
			{Key: "foreignField", Value: "id"},
			{Key: "pipeline",
				Value: bson.A{
					bson.D{
						{Key: "$project",
							Value: bson.D{
								{Key: "password", Value: 0},
								{Key: "_id", Value: 0},
								{Key: "id", Value: 0},
								{Key: "created_at", Value: 0},
								{Key: "updated_at", Value: 0},
							},
						},
					},
				},
			},
			{Key: "as", Value: "user"},
		}}},
	}

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
	res.Status = "success"
	rest.RenderJSON(w, res)
}

// A request on Team endpoint returns all information about all Teams

//     Http Method : GET http://chowie.uk/api/v1/team
//     Http Method : GET http://localhost:8080/api/v1/team

func allTeamsResponseHandler(w http.ResponseWriter, r *http.Request) {

	pipeline := mongo.Pipeline{bson.D{
		{Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "user_id"},
				{Key: "foreignField", Value: "id"},
				{Key: "pipeline",
					Value: bson.A{
						bson.D{
							{Key: "$project",
								Value: bson.D{
									{Key: "password", Value: 0},
									{Key: "_id", Value: 0},
									{Key: "id", Value: 0},
									{Key: "created_at", Value: 0},
									{Key: "updated_at", Value: 0},
								},
							},
						},
					},
				},
				{Key: "as", Value: "user"},
			},
		},
	}}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := teamCollection.Aggregate(ctx, pipeline)
	defer cancel()
	if err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to fetch teams")
	}

	var res entity.TeamResponse

	if err = cursor.All(ctx, &res.Teams); err != nil {
		rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse teams")
	}
	defer cancel()
	res.Status = "success"

	rest.RenderJSON(w, res)
}
