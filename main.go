package main

import (
	"context"
	"encoding/json"
	"fmt"
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
		Debug:            true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	router.Post("/register", registrationHandler) // post registration route
	router.Group(func(r chi.Router) {
		r.Use(m.Auth)
		r.Use(m.UpdateUser(middleware.UserUpdFunc(func(user token.User) token.User {
			user.SetStrAttr("some_attribute", "attribute value")
			return user
		})))
		r.Get("/private_data", protectedDataHandler) // protected api
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

	if err != nil {
		// defer cancel() adding a defer here fixes highlighted context problem
		// not sure about ramnificatiosn of doing so
		rest.SendErrorJSON(w, r, log.Default(), http.StatusBadRequest, err, "failed to parse registration")
		return
	}

	log.Printf("[DEBUG] checking if provided email exists in mongodb")

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

	password := service.HashPassword(user.Password)
	user.Password = password

	user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	userCount, _ := userCollection.CountDocuments(ctx, bson.D{})
	user.User_id = fmt.Sprint(userCount + 1)

	err = allocateTeam(&user, ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "all teams have been allocated")
			return
		}
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

func allocateTeam(user *entity.User, ctx context.Context) error {

	var team entity.TeamData

	// Using Aggregation / samples

	nextAvailableTeam, err := teamCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "user_id", Value: nil}}}},
		bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	})

	nextAvailableTeam.Decode(&team)

	// Using Find.
	// cursor, err := teamCollection.Find(
	// 	ctx,
	// 	bson.D{
	// 		{"user_id", nil},
	// 	})

	// Using FindOne

	// err := teamCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: nil}}).Decode(&team)

	if err != nil {
		log.Printf("[DEBUG] failed when attempting to find an available team")
		return err
	}

	teamCollection.UpdateByID(ctx, team.ID, bson.D{{Key: "user_id", Value: user.User_id}})

	if err != nil {
		log.Printf("[DEBUG] failed when attempting to update team %s (Object ID: %s) with user id %s ", team.Name, team.ID, user.User_id)
		return err
	}

	user.Team_id = team.Team_id
	log.Printf("[INFO] successfully allocated %s (id %s) to %s (id %s)\n", team.Name, team.Team_id, user.Full_Name, user.User_id)
	// log.Printf("[INFO] available teams left: %s)\n",)
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
