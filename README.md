# Sweepstakes-app

## Building

### Frontend

`npm install --production`

`npm run build`

### Backend

Populate `.env.production.local`
Current version includes:

```
MONGO_DB_URL=
MONGO_DB_PASSWORD=
MONGO_DB_NAME=
MONGO_COLLECTION_NAME=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
FACEBOOK_APP_ID=
FACEBOOK_APP_SECRET=
SMTP_USER=
SMTP_PASS=
SECRET_KEY=
```

`go build -o server`
 
## Run Server

`./server`

Server should now be running on http://localhost:8080

## Authorization Examples

Head to http://localhost:8080/web/ for an example of the current authorization methods

You can find the front end code in `/auth-example-frontend/main.js`

## Routes

Open routes:

`/web`, `/register`, `/auth`, `/avatar`

Protected routes:

`/private_data` 

## API Specification

### Registration

You can register to a mongodb store with a POST request to `/register` with a body such as:

```
{
    "username": "user",
    "password": "secret"
}
```

### Authorization

Authorization handlers can be access via POST requests to `/auth/<handler>`

#### MongoDB

You can login via a check against a mongodb store with a POST request in the form
`http://localhost:8080/auth/mongo/login?id=sweepstakes&user=<username>&passwd=<password>`

Using a valid username and password combo (such as `test` & `secret`) will give you a response including your name, id and avatar, and set both a JWT and XSRF-TOKEN

#### Anonymous

You can login with any username with a POST request in the form
`http://localhost:8080/auth/anonymous/login?id=sweepstakes&user=<username>`

You will again receive a response including your name, id and avatar, and set both a JWT and XSRF-TOKEN, but clients with ids prepended with `anonymous` are not permitted along the protected routes

#### oauth2 - Google & Facebook

You can only use these providers if you've set the relevant environment variables, and the server is deployed on a domain that has been registered and verified / gone through the steps required.

#### oauth2 - dev

For *development purposes only* an dummy oauth2 dev server is provided. You can login with any username. Notice that you'll get more info if you login with `dev_admin` as the username.

## Data Model

Here's the current model for the User entity. At this time I'm not handling email, just username and password

// User is a reduced model of objects that will be retrived or inserted into the DB
```
type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Username   string             `json:"Username"`
	Password   string             `json:"Password"`
	User_id    string             `json:"user_id"`
	//Email      string             `json:"email"`
	//Nation     string             `json:"nation"`
}
```

I have yet to add a data model for Nations.

## TODO

- Add Nations data model
- Add logic for allocating nations
- Start building out required features
- Consider local copy of World Cup 2022 API, + cronjob to populate the database within rate limits
- Consider rate limiting / throttling:
	- registration attempts
	- login attempts
