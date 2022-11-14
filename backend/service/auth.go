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
//var Client *mongo.Client = repo.DBinstance()

//var userCollection *mongo.Collection = repo.OpenCollection(Client, "users")
//var teamCollection *mongo.Collection = repo.OpenCollection(Client, "teams")

func InitializeAuth(userCollection *mongo.Collection, teamCollection *mongo.Collection) *auth.Service {

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

				if strings.HasPrefix(claims.User.ID, "mongo_") {
					err := repo.SwapEmailNameClaims(userCollection, &claims)

					if err != nil {
						if err == mongo.ErrNoDocuments {
							log.Printf("[INFO] no user with that ID found.")
						}
						if err != nil {
							log.Printf("[DEBUG] error swapping claims names %v", err)
						}
					}
				}

				// Check if the user is in the db
				inDb, err := repo.UserInCollection(userCollection, *claims.User)
				if err != nil {
					log.Printf("[DEBUG] error checking if user exists in db: %v", err)
				}
				if !inDb {
					log.Printf("[INFO] user doesn't exist exist in db. Adding user.")
					// Non social login users must be in the db
					err = repo.AddSocialUser(userCollection, *claims.User)
					if err != nil {
						log.Printf("[DEBUG] failed adding social user to db: %v", err)
					}

				}

				// Check if the user has been assigned a team
				// As of writing only social login users can login without being assigned a team
				team, err := repo.GetUserTeam(teamCollection, *claims.User)
				if err != nil {
					if err == mongo.ErrNoDocuments {
						log.Printf("[INFO] no team assigned to user. Attempting to allocate team")
						ok, err := repo.CheckTeamAvailability(teamCollection)
						if err != nil {
							log.Printf("[DEBUG] error checking availability")
						}
						if ok {
							team, err = repo.AllocateTeamSocial(teamCollection, *claims.User)
							if err != nil {
								log.Printf("[DEBUG] error allocating social team: ", err)
							}
							err = repo.UpdateSocialUserWithTeam(userCollection, *claims.User, team)
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
	service.AddDirectProvider("mongo", mongoAuthProvider(userCollection))

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
