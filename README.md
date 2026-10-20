# Chirpy

Chirpy is the backend for a small Twitter-style social network. Users sign up, log in, and post short messages called chirps. It's a JSON HTTP API written in Go with PostgreSQL behind it.

## Why it's worth a look

Chirpy is small enough to read in an afternoon, but it covers most of what a real API server needs:

- **Routing with only the standard library.** It uses Go's `net/http` `ServeMux`, including method-based routes like `GET /api/chirps/{chirpID}`, with no web framework.
- **Type-safe SQL.** Queries are plain SQL, and [sqlc](https://sqlc.dev) generates the Go code that runs them. Schema changes are versioned migrations run with [goose](https://github.com/pressly/goose).
- **Real authentication.** Passwords are hashed with Argon2id. Login returns a one-hour JWT access token and a 60-day refresh token that's stored in the database and can be revoked.
- **Authorization.** You need to be logged in to post or delete a chirp, and you can only delete your own.
- **Webhooks.** A payment provider ("Polka") can upgrade a user to Chirpy Red by calling a webhook protected with an API key.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26 or newer
- [PostgreSQL](https://www.postgresql.org/download/) 15 or newer, running locally
- goose, for migrations:
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
- sqlc (only needed if you change the SQL in `sql/queries`):
  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

## Setup

1. **Clone the repository.**
   ```bash
   git clone https://github.com/Dannybouy/chirpy.git
   cd chirpy
   ```

2. **Create a database.**
   ```bash
   createdb chirpy
   ```

3. **Create a `.env` file** in the project root:
   ```env
   DB_URL="postgres://<user>:<password>@localhost:5432/chirpy?sslmode=disable"
   PLATFORM="dev"
   JWT_SECRET="<random string>"
   POLKA_KEY="<any string; webhook callers must send it>"
   ```
   Generate a JWT secret with `openssl rand -base64 64`. The server refuses to start if `DB_URL`, `JWT_SECRET`, or `POLKA_KEY` is missing. Setting `PLATFORM` to `dev` enables `POST /admin/reset`, which deletes all users. Don't set it to `dev` anywhere but your own machine.

4. **Run the migrations.**
   ```bash
   cd sql/schema
   goose postgres "postgres://<user>:<password>@localhost:5432/chirpy" up
   cd ../..
   ```

5. **Build and start the server.**
   ```bash
   go build -o chirpy && ./chirpy
   ```
   The server listens on `http://localhost:8080`. Check that it's up with:
   ```bash
   curl http://localhost:8080/api/healthz
   ```

## API

Endpoints marked **JWT** need an `Authorization: Bearer <access token>` header. The refresh and revoke endpoints take the refresh token in that header instead.

### Users and auth

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/users` | None | Create a user from `{"email", "password"}` |
| `PUT` | `/api/users` | JWT | Change the logged-in user's email and password |
| `POST` | `/api/login` | None | Log in; returns a `token` and a `refresh_token` |
| `POST` | `/api/refresh` | Refresh token | Get a new one-hour access token |
| `POST` | `/api/revoke` | Refresh token | Revoke a refresh token |

### Chirps

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/chirps` | JWT | Post a chirp from `{"body"}` (140 characters max) |
| `GET` | `/api/chirps` | None | List chirps. Optional `?author_id=<uuid>` and `?sort=asc` or `?sort=desc` (by creation time, ascending by default) |
| `GET` | `/api/chirps/{chirpID}` | None | Get one chirp |
| `DELETE` | `/api/chirps/{chirpID}` | JWT | Delete one of your own chirps |

### Other

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/healthz` | None | Readiness check; returns `OK` |
| `POST` | `/api/polka/webhooks` | `Authorization: ApiKey <POLKA_KEY>` | Handles `user.upgraded` events and marks the user as Chirpy Red |
| `GET` | `/admin/metrics` | None | HTML page showing how many times `/app` was visited |
| `POST` | `/admin/reset` | `PLATFORM=dev` only | Delete all users and reset the visit counter |
| `GET` | `/app/` | None | Static files from the project root |

### Try it

```bash
curl -X POST localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","password":"hunter2"}'

TOKEN=$(curl -s -X POST localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","password":"hunter2"}' | jq -r .token)

curl -X POST localhost:8080/api/chirps \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"body":"Hello, Chirpy!"}'

curl localhost:8080/api/chirps
```

The examples use [`jq`](https://jqlang.github.io/jq/) to pull the token out of the login response.

## Development

- Run the tests with `go test ./...`. Tests currently cover the `internal/auth` package: password hashing, creating and validating JWTs, and reading the `Authorization` header.
- After editing anything in `sql/queries`, regenerate the Go code with `sqlc generate` from the project root. Don't edit `internal/database` by hand.
- To change the schema, add a new numbered file in `sql/schema` rather than editing one that has already been applied.

## Project layout

```
main.go              server setup and route registration
chirps.go            chirp handlers
users.go             signup, login, and user update handlers
tokens.go            refresh and revoke handlers
webhooks.go          Polka webhook handler
internal/auth/       password hashing, JWTs, header parsing
internal/database/   code generated by sqlc (don't edit)
sql/schema/          goose migrations
sql/queries/         SQL that sqlc compiles into Go
```
