# chirpy

boot.dev lesson

This lesson covers http server application creation

It's barebones Twitter type server application.
It's just created for learning experience and not intended for actual solution.

## Additional tools used

- postgresql is used as db

- goose was used as db migration tool

- sqlc is used to generate go code that interacts with db

## Setup requirements

application is witten with go version

application root/start directory needs .env file with follwing values.
```bash
DB_URL="" #database connection url that application uses
PLATFORM="" #if dev then /admin/reset endpoint is allowed to be used to clear database
CHIRPY_SECRET="" #Secret used for generating JWT
POLKA_KEY="" #API key we know to trust for webhook from Polka payment system
```

Those variables are handled as system environment variables.



## REST endpiont documentation

#### /api/healthz

Request Type: GET

standard readiness endpoint

#### /api/chirps", api_cfg.handlerGetAllChirps)

Request Type: **GET**

Returns all chirps in creation date ascending order.

Accespts optional url query parameter '?author_id=' to get specific authors chiprs

Request Type: **POST**

Stores the new chirp to the database, example body:
```json
{
  "body": "chirp content"
}
```

#### /api/chirps/{chirpID}

Request Type: **GET**

Retreives the chirp with given uudi

Request Type: **DELETE**

Deletes the chirp with given uudi.

Only the author is allowed to delete chirp.

#### /api/users

Request Type: **POST**

Registers new user to the system, example body:
```json
{
  "email": "example@example.ex",
  "password": "password1234"
}
```

Request Type: **PUT**

Allows user to change their email and password, example body:
```json
{
  "email": "new_example@example.ex",
  "password": "new_password1234"
}
```

#### /api/login

Request Type: **POST**

Logs in user to the system and returns JWT and refresh token, example body:
```json
{
  "email": "example@example.ex",
  "password": "password1234"
}
```

#### /api/refresh 

Request Type: **POST**

Rerfresh token needs to be sent in header Authorization parameter and new JWT is returned to continue the session.

#### /api/refresh

Request Type: **POST**

Refresh token needs to be sent in header Authorization parameter and the refresh token is revoked/invalidated

#### /api/polka/webhooks

Webhook endpoint for Polka payment system to send confirmations for user payments so they can be upgraded to Chirpy Red status

#### /admin/reset

Request Type: **POST**

**Only available in Dev envionment**

Deletes all content from database.

## Improvement area notes

There are some TODO comments in the code to review and improve. Generally for better responses to API calls that fail for some reason.

DRY up the code, nr1 example is to have json replies for success and failure as separate functions, currenty the same type of code is constalty used over and over everywhere.

Review and clean up some response logic.

**NB!** There is no real plan to make any significante improvements and additional development on this project, maybe only smaller improvements for additional learning.
