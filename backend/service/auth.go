package service

import (
	"context"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chowieuk/sweepstakes-app/backend/entity"
	"github.com/chowieuk/sweepstakes-app/backend/repo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/provider"
	"github.com/go-pkgz/auth/token"
	log "github.com/go-pkgz/lgr"
	//"github.com/go-pkgz/auth/provider/sender" Sender was giving me go mod issues
)

// Client Database instance
var Client *mongo.Client = repo.DBinstance()

var userCollection *mongo.Collection = repo.OpenCollection(Client, "users")
var teamCollection *mongo.Collection = repo.OpenCollection(Client, "teams")

func InitializeAuth(collection *mongo.Collection) *auth.Service {

	// define auth options
	options := auth.Opts{
		SecretReader: token.SecretFunc(func(_ string) (string, error) { // secret key for JWT, ignores aud
			return "secret", nil // TODO: Research and potentially adjust to a ENV variable
		}),
		TokenDuration:     time.Minute,                                 // short token, refreshed automatically
		CookieDuration:    time.Hour * 24,                              // cookie fine to keep for long time
		DisableXSRF:       true,                                        // don't disable XSRF in real-life applications!
		Issuer:            "PaChowie Sweepstakes",                      // part of token, just informational
		URL:               "http://localhost:8080",                     // base url of the protected service
		AvatarStore:       avatar.NewLocalFS("/tmp/demo-auth-service"), // stores avatars locally
		AvatarResizeLimit: 200,                                         // resizes avatars to 200x200
		ClaimsUpd: token.ClaimsUpdFunc(func(claims token.Claims) token.Claims { // modify issued token
			if claims.User != nil && claims.User.Name == "dev_admin" { // set attributes for dev_admin
				claims.User.SetAdmin(true)
				claims.User.SetStrAttr("custom-key", "some value")
			} else if claims.User != nil {

				// Check if the user is in the db
				inDb, err := UserInDb(*claims.User)
				if err != nil {
					log.Printf("[DEBUG] error checking if user exists in db: %v", err)
				}
				if !inDb {
					log.Printf("[INFO] user doesn't exist exist in db. Adding user.")
					// Non social login users must be in the db
					err = AddSocialUser(*claims.User)
					if err != nil {
						log.Printf("[DEBUG] failed adding social user to db: %v", err)
					}

				}

				// Check if the user has been assigned a team
				// As of writing only social login users can login without being assigned a team
				team, err := GetUserTeam(*claims.User)
				if err != nil {
					if err == mongo.ErrNoDocuments {
						log.Printf("[INFO] no team assigned to user. Attempting to allocate team")
						ok, err := CheckTeamAvailability()
						if err != nil {
							log.Printf("[DEBUG] error checking availability")
						}
						if ok {
							team, err = AllocateTeamSocial(*claims.User)
							if err != nil {
								log.Printf("[DEBUG] error allocating social team: ", err)
							}
							err = UpdateSocialUserWithTeam(*claims.User, team)
							if err != nil {
								log.Printf("[DEBUG] error updating social user with team", err)
							}
						}
					}
				}
				if team.Name != "" {
					claims.User.SetStrAttr("team_name", team.Name)
					claims.User.SetStrAttr("team_flag", team.Flag)
				}
				if team.Name == "" {
					claims.User.SetStrAttr("team_name", "Waiting List")
				}
			}
			return claims
		}),
		Validator: token.ValidatorFunc(func(_ string, claims token.Claims) bool { // rejects some tokens
			if claims.User != nil {

				if strings.HasPrefix(claims.User.ID, "google_") { // allow all users with google auth
					return true
				}
				if strings.HasPrefix(claims.User.ID, "facebook_") { // allow all users with facebook auth
					return true
				}
				if strings.HasPrefix(claims.User.ID, "mongo_") { // allow all users with mongo auth
					return true
				}
				if strings.HasPrefix(claims.User.ID, "dev_") { // allow all users with dev auth
					return true
				}
				// if strings.HasPrefix(claims.User.Name, "dev_") { // only dev_* names are permitted
				// 	return true
				// }
			}
			return false
		}),
		Logger:      log.Default(), // optional logger for auth library
		UseGravatar: true,          // for verified provider use gravatar service
	}

	// create auth service
	service := auth.NewService(options)
	service.AddProvider("dev", "", "")                                                              // add dev provider
	service.AddProvider("google", os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET")) // add google provider
	service.AddProvider("facebook", os.Getenv("FACEBOOK_APP_ID"), os.Getenv("FACEBOOK_APP_SECRET")) // add facebook provider

	// allow anonymous user via custom (direct) provider
	service.AddDirectProvider("anonymous", anonymousAuthProvider())

	// allow checking credentials via mongodb store
	service.AddDirectProvider("mongo", mongoAuthProvider(collection))

	// namecheap email sender setup
	// namecheapSender := sender.NewEmailClient(sender.EmailParams{
	// 	Host:         "mail.privateemail.com",
	// 	Port:         567,
	// 	SMTPUserName: "info@chowie.uk",
	// 	SMTPPassword: os.Getenv("SMTP_PASS"),
	// 	TLS:          true,
	// 	From:         "A PaChowie Endeavour",
	// 	Subject:      "Chowie sent you some email!",
	// 	ContentType:  "text/html",
	// 	//Charset:      "UTF-8",
	// }, log.Default())

	// // Note: This email template can be HTML
	// // TODO: Build some email templates
	// msgTemplate := "Hi {{.User}}, here's your confirmation email!\n To confirm please follow http://chowie.uk/auth/email/login?token={{.Token}}\n{{.Address}}\n{{.Site}}"
	// service.AddVerifProvider("email", msgTemplate, namecheapSender)

	// run dev/test oauth2 server on :8084
	go func() {
		devAuthServer, err := service.DevAuth() // peak dev oauth2 server
		if err != nil {
			log.Printf("[PANIC] failed to start dev oauth2 server, %v", err)
		}
		devAuthServer.Run(context.Background())
	}()

	return service
}

// monoAuthProvider checks credentials against a mongodb store
func mongoAuthProvider(collection *mongo.Collection) provider.CredCheckerFunc {
	log.Printf("[DEBUG] mongo provider enabled")
	return (func(user, password string) (ok bool, err error) {
		ok, err = checkMongo(collection, user, password)
		return ok, err
	})
}

// checkMongo compares given credentials with our mongodb collection
func checkMongo(collection *mongo.Collection, user, password string) (ok bool, err error) {
	log.Printf("[INFO] checking provided credentials against mongodb")
	if user == "" {
		log.Printf("[DEBUG] user not provided: %s", user)
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var foundUser entity.User
	err = collection.FindOne(ctx, bson.M{"email": user}).Decode(&foundUser)

	defer cancel()
	if err != nil {
		log.Printf("[DEBUG] %s not found in collection %s", user, collection.Name())
		return false, err
	}

	passwordIsValid := CheckPasswordHash(password, *foundUser.Password)
	defer cancel()
	if !passwordIsValid {
		log.Printf("[DEBUG] password does not match hash found in collection %s\n", collection.Name())
		return false, err
	}
	return true, nil
}

// HashPassword generates a bcrypt has from a provided password
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("[ERROR] Error computing hash", err)
	}
	return string(bytes)
}

// CheckPasswordHash compares a bcrypt hashed password with its possible plaintext equivalent. Returns true on success
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// anonymousAuthProvider allows auth-free login with any valid user name
func anonymousAuthProvider() provider.CredCheckerFunc {
	log.Printf("[WARN] anonymous access enabled")
	var isValidAnonName = regexp.MustCompile(`^[a-zA-Z][\w ]+$`).MatchString

	return func(user, _ string) (ok bool, err error) {
		user = strings.TrimSpace(user)
		if len(user) < 3 {
			log.Printf("[WARN] name %q is too short, should be at least 3 characters", user)
			return false, nil
		}

		if !isValidAnonName(user) {
			log.Printf("[WARN] name %q should have letters, digits, underscores and spaces only", user)
			return false, nil
		}
		return true, nil
	}
}

// UserInDb checks if the user exists in mongodb
func UserInDb(user token.User) (bool, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	log.Printf("[INFO] checking if user %v (id %v) exists in mongodb", user.Name, user.ID)

	count, err := userCollection.CountDocuments(ctx, bson.M{"id": user.ID})
	// log.Printf("[DEBUG] number of records for user %v (id %v) = %v", user.Name, user.ID, count)
	defer cancel()
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// AddSocialUser Adds a social login user (Google or Facebook) to our users collection
func AddSocialUser(user token.User) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var newUser entity.SocialUser
	newUser.ObjectID = primitive.NewObjectID()
	newUser.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	newUser.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	newUser.Full_Name = user.Name
	newUser.Email = user.Email
	newUser.ID = user.ID

	resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, newUser)
	defer cancel()
	if insertErr != nil {
		return insertErr
	}
	log.Printf("[INFO] successfully added %s to mongodb %s", newUser.Full_Name, resultInsertionNumber)
	return nil
}

// GetUserTeam provides the team associated with a token user
func GetUserTeam(user token.User) (entity.TeamData, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var team entity.TeamData
	// attempt to find a team assigned to the user
	err := teamCollection.FindOne(ctx, bson.M{"user_id": user.ID}).Decode(&team)
	defer cancel()
	if err != nil {
		log.Printf("[DEBUG] failed when attempting to find an available team")
		return entity.TeamData{}, err
	}
	return team, nil
}

// CheckTeamAvailability returns true if a team is available
func CheckTeamAvailability() (bool, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	log.Printf("[DEBUG] checking amount of available teams")

	count, err := teamCollection.CountDocuments(ctx, bson.M{"user_id": primitive.Null{}})
	defer cancel()
	if err != nil {
		return false, err
	}
	if count > 0 {
		log.Printf("[DEBUG] number of teams available: %v", count)
		return true, nil
	}
	log.Printf("[DEBUG] !!! NO TEAMS AVAILABLE !!!")
	return false, nil
}

// AllocateTeamSocial assigns a team to a Social user
func AllocateTeamSocial(user token.User) (entity.TeamData, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var team entity.TeamData

	log.Printf("[INFO] attempting to allocate team to social user")

	// Note: This should only be executed after a check for team availability
	team, err := repo.RandomUnassignedTeam(teamCollection, context.TODO())
	defer cancel()
	if err != nil {
		log.Printf("[DEBUG] failed when attempting to find an available team")
		return team, err
	}

	result, err := teamCollection.UpdateByID(ctx, team.ObjectID.Hex(), bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "user_id",
			Value: user.ID}}}})
	defer cancel()
	if !(result.ModifiedCount > 0) {
		log.Printf("[DEBUG] no team documents were modified")
	}

	if err != nil {
		log.Printf("[DEBUG] failed when attempting to update team %s (Object ID: %s) with user %s id %s ", team.Name, team.ID, user.Name, user.ID)
		return entity.TeamData{}, err
	}

	user.SetStrAttr("team_name", team.Name)
	user.SetStrAttr("team_flag", team.Flag)
	log.Printf("[INFO] successfully allocated %s (id %s) to %s (id %s)\n", team.Name, team.ID, user.Name, user.ID)
	// log.Printf("[INFO] available teams left: %s)\n",)
	return team, nil
}

func UpdateSocialUserWithTeam(user token.User, team entity.TeamData) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	log.Printf("[DEBUG] updating social user entry in db with team id")

	result, err := userCollection.UpdateOne(ctx, bson.D{{Key: "id", Value: user.ID}}, bson.D{{
		Key: "$set",
		Value: bson.D{{
			Key:   "team_id",
			Value: team.ID}}}})
	defer cancel()
	if !(result.ModifiedCount > 0) {
		log.Printf("[DEBUG] no records were modified")
		log.Printf("[DEBUG] no record added when attempting to update user %s (ID: %s) with team %s (ID: %s) ", user.Email, user.ID, team.Name, team.ID)
	}

	if err != nil {
		log.Printf("[DEBUG] failed when attempting to update user %s (ID: %s) with team %s (ID: %s) ", user.Email, user.ID, team.Name, team.ID)
		return err
	}
	return nil
}
