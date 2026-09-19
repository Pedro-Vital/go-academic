# Capstone: A REST API with Gin, SQLite, and JWT

> Every guide in `notes/` up to this point taught one slice of Go in isolation — structs and methods, interfaces, error handling, pointers, goroutines, modules, JSON. This project is where those slices get assembled into something that actually looks like backend work: an HTTP API, backed by a real (if tiny) database, that creates accounts, issues auth tokens, and lets authenticated users manage a resource. If you've built anything with FastAPI + SQLAlchemy + `python-jose`, this project is the direct structural equivalent — except every piece that a Python framework would wire up for you implicitly (routing, request parsing, dependency injection, auth) is instead something you can point to a specific line of Go for. That's the whole trade-off this guide keeps coming back to: more visible plumbing, in exchange for nothing happening by magic.

This README treats the project as a finished artifact to *read*, not a tutorial to follow step by step — it walks through every file, explains why it's shaped the way it is, and calls out a few real bugs left in on purpose as spot-the-bug exercises.

## Table of Contents

1. [What This Project Does](#1-what-this-project-does)
2. [Project Structure](#2-project-structure)
3. [Dependencies: Reading `go.mod`](#3-dependencies-reading-gomod)
4. [Application Bootstrap: `main.go`](#4-application-bootstrap-maingo)
5. [The Database Layer: `db/db.go`](#5-the-database-layer-dbdbgo)
6. [Models: Hand-Written SQL Instead of an ORM](#6-models-hand-written-sql-instead-of-an-orm)
7. [Routing with Gin: `routes/routes.go`](#7-routing-with-gin-routesroutesgo)
8. [Route Handlers: Events, Signup, Login, Registration](#8-route-handlers-events-signup-login-registration)
9. [Password Hashing: `utils/hash.go`](#9-password-hashing-utilshashgo)
10. [JWT Authentication: `utils/jwt.go`](#10-jwt-authentication-utilsjwtgo)
11. [The Auth Middleware: `middlewares/auth.go`](#11-the-auth-middleware-middlewaresauthgo)
12. [Full Request Walkthrough: Creating an Event](#12-full-request-walkthrough-creating-an-event)
13. [Manual Testing: the `api-test/*.http` Files](#13-manual-testing-the-api-testhttp-files)
14. [Running the Project](#14-running-the-project)
15. [Rough Edges: Bugs and Exercises](#15-rough-edges-bugs-and-exercises)
16. [Key Takeaways](#16-key-takeaways)

---

## 1. What This Project Does

It's a small **event management API**: users sign up, log in, and (once authenticated) create/update/delete events and register/unregister for them. Concretely:

| Method | Path | Auth required | Purpose |
|---|---|---|---|
| `POST` | `/signup` | No | Create a user account |
| `POST` | `/login` | No | Authenticate, receive a JWT |
| `GET` | `/events` | No | List all events |
| `GET` | `/events/:id` | No | Fetch one event |
| `POST` | `/events` | Yes | Create an event (owned by the caller) |
| `PUT` | `/events/:id` | Yes, owner only | Update an event |
| `DELETE` | `/events/:id` | Yes, owner only | Delete an event |
| `POST` | `/events/:id/register` | Yes | Register the caller for an event |
| `DELETE` | `/events/:id/register` | Yes | Cancel the caller's registration |

Three SQLite tables back this: `users`, `events`, and a join table `registrations` connecting users to the events they've signed up for (a classic many-to-many relationship — one user can register for many events, one event can have many registered users).

---

## 2. Project Structure

```
11-REST-API/
├── main.go              # entry point: wires DB + routes, starts the server
├── go.mod / go.sum       # module definition and dependency lock (Guide 8)
├── db/
│   └── db.go             # opens the SQLite connection, creates tables
├── models/
│   ├── event.go          # Event struct + all event-related SQL
│   └── user.go           # User struct + signup/login SQL
├── routes/
│   ├── routes.go         # maps HTTP method + path -> handler function
│   ├── events.go         # handlers for /events*
│   ├── register.go       # handlers for /events/:id/register
│   └── users.go          # handlers for /signup, /login
├── middlewares/
│   └── auth.go           # JWT-checking Gin middleware
├── utils/
│   ├── hash.go           # bcrypt password hashing
│   └── jwt.go            # JWT generation + verification
└── api-test/
    └── *.http            # example requests for a REST-client editor plugin
```

This is a layered structure you'll recognize from any Django or FastAPI project, just with Go-flavored names: `models/` is your ORM-models-and-queries layer, `routes/` is your `views.py`/`routers/` layer, `middlewares/` is Django middleware / FastAPI dependencies, and `utils/` is the everything-else drawer every project accumulates. Guide 1 called this "flat and explicit" package layout out — nothing here is auto-discovered or convention-magic'd into place; `main.go` explicitly imports and calls into every layer, in order.

---

## 3. Dependencies: Reading `go.mod`

```go
module example.com/rest-api

go 1.21.2

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/mattn/go-sqlite3 v1.14.17
	golang.org/x/crypto v0.14.0
)
```

Four direct dependencies, each doing one job — this project deliberately avoids a kitchen-sink framework:

- **`gin-gonic/gin`** — the HTTP router and request/response framework. Go's standard library `net/http` (Guide 7's territory) is perfectly usable directly, but Gin adds route parameters (`:id`), route grouping, middleware chaining, and JSON binding/validation on top — roughly what Flask or FastAPI's routing layer gives you over raw WSGI/ASGI.
- **`golang-jwt/jwt/v5`** — encodes/decodes and signs/verifies JSON Web Tokens. The Python equivalent is `python-jose` or `PyJWT`.
- **`mattn/go-sqlite3`** — a `database/sql` driver for SQLite, written as a CGo binding to the real SQLite C library. Equivalent to Python's built-in `sqlite3` module.
- **`golang.org/x/crypto`** — the extended (non-stdlib-but-official) crypto package; this project uses only its `bcrypt` subpackage. Equivalent to Python's `bcrypt` or `passlib`.

Everything else in `go.mod`'s second `require` block is marked `// indirect` — dependencies of dependencies (Gin pulls in a JSON encoder, a validator, etc.) that Go's tooling tracks automatically but that this project's own code never imports directly. This is the same distinction `go.sum`(the full transitive lock file) makes precise, covered in Guide 8.

One quietly important line lives in `db/db.go`, not `go.mod`:

```go
import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)
```

The **blank import** (`_ "github.com/..."`) is a Go idiom for "run this package's `init()` for its side effects, but I never call anything on it by name." `go-sqlite3`'s `init()` registers itself with `database/sql` as a driver named `"sqlite3"` — so `sql.Open("sqlite3", ...)` later in the same file can find it. This is `database/sql`'s whole design: it's an *interface* (Guide 4) that any driver can plug into, and the blank import is how a driver gets plugged in without the calling code needing a direct reference to any of the driver's exported types. Python's DB-API 2.0 drivers (`psycopg2`, `sqlite3`) work on a similar "common interface, swappable implementation" principle, just without needing an explicit side-effect-only import to register.

---

## 4. Application Bootstrap: `main.go`

The entire entry point is 16 lines:

```go
package main

import (
	"example.com/rest-api/db"
	"example.com/rest-api/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()
	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":8080") // localhost:8080
}
```

Three steps, in a strict, explicit order: open the database, build a Gin engine, attach every route to it, then block forever serving HTTP on port 8080. `gin.Default()` (as opposed to the bare `gin.New()`) pre-attaches two pieces of middleware to every request: a request logger and a panic-recovery handler that turns an unhandled panic into a `500` instead of crashing the whole process — cheap insurance you'd otherwise have to write yourself.

Note the import paths: `example.com/rest-api/db` and `example.com/rest-api/routes` are **not** real internet domains — `example.com/rest-api` is just the module name declared in `go.mod`'s first line, and every subdirectory is importable as `<module-name>/<subdirectory>`, exactly the way `go.mod`'s Guide 8 material describes for any local multi-package project.

---

## 5. The Database Layer: `db/db.go`

```go
var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "api.db")

	if err != nil {
		panic("Could not connect to database.")
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	createTables()
}
```

`DB` is a **package-level variable** — every other package that needs to run a query imports `example.com/rest-api/db` and reads `db.DB` directly, rather than the database handle being passed as an explicit argument through every function call. This is a deliberate simplicity trade-off appropriate for a small capstone (no dependency injection, no interfaces to swap out a real DB for a test double), but it is a trade-off: it makes `models/` implicitly coupled to a single global connection, which is exactly the kind of thing Guide 8's testing material would flag as harder to unit-test in isolation than a version where each model function took a `*sql.DB` parameter.

`sql.Open` here does **not** actually open a network connection or verify the file is readable — it just validates the driver name and prepares a connection pool that connects lazily on first use. That's why `panic` on its error here is really only catching a malformed driver name, not "the database file doesn't exist" (SQLite creates `api.db` on disk automatically if it isn't there yet, which is part of why SQLite is such a convenient choice for a learning project — zero setup, no separate server process to run, just a file).

`createTables()` then runs three `CREATE TABLE IF NOT EXISTS` statements — `users`, `events`, `registrations` — every time the app starts. This is a crude but effective substitute for the migration tooling a larger project would use (Django's `migrate`, Alembic, `golang-migrate`): idempotent schema setup baked directly into startup, fine for a single-developer capstone, not something you'd want at production scale where schema changes need to be versioned and reviewed independently of app deploys.

The schema itself:

```sql
CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL
)

CREATE TABLE events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	location TEXT NOT NULL,
	dateTime DATETIME NOT NULL,
	user_id INTEGER,
	FOREIGN KEY(user_id) REFERENCES users(id)
)

CREATE TABLE registrations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	event_id INTEGER,
	user_id INTEGER,
	FOREIGN KEY(event_id) REFERENCES events(id),
	FOREIGN KEY(user_id) REFERENCES users(id)
)
```

Textbook relational modeling: `events.user_id` is a one-to-many foreign key (one user owns many events), and `registrations` is the join table making user↔event a many-to-many relationship. `users.password` stores a bcrypt hash (Section 9), never a plaintext password — the column name is a little misleading on its own, but the code that writes to it (`models/user.go`) always hashes first.

---

## 6. Models: Hand-Written SQL Instead of an ORM

Coming from Django's ORM or SQLAlchemy, the biggest adjustment in `models/event.go` and `models/user.go` is that there is no ORM at all here. Every query is a hand-written SQL string executed through `database/sql`'s three core primitives, and every result is manually `Scan`'d back into struct fields. This is normal, idiomatic Go for a project this size — heavier projects reach for a query builder (`sqlx`, `squirrel`) or a code-generating ORM (`gorm`, `sqlc`), but plain `database/sql` is still a completely standard way to talk to a relational database in Go.

### 6.1 The `Event` struct and its `binding` tags

```go
type Event struct {
	ID          int64
	Name        string    `binding:"required"`
	Description string    `binding:"required"`
	Location    string    `binding:"required"`
	DateTime    time.Time `binding:"required"`
	UserID      int64
}
```

This is the same struct-tag mechanic Guide 9 covered for `validate:"..."` — Gin's `ShouldBindJSON` (Section 8) uses `go-playground/validator` internally too, it just reads the tag under the name `binding` instead of `validate`, since Gin's own documentation and ecosystem settled on that name before `validate` became the more common standalone convention. Functionally identical: `binding:"required"` rejects a request whose JSON is missing that field (or supplies its zero value) before the handler function's body even runs.

One field notably has **no** `binding` tag: `UserID`. That's intentional — the client never supplies `UserID` in the request body; the handler fills it in from the authenticated user's ID after binding succeeds (Section 8.3). If `UserID` had a `required` tag, every legitimate request would fail validation, since the field is deliberately absent from client input.

### 6.2 CRUD methods, and a receiver-type bug worth noticing

```go
func (e *Event) Save() error {
	query := `INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)`
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(e.Name, e.Description, e.Location, e.DateTime, e.UserID)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	e.ID = id
	return err
}
```

`Save` uses a **pointer receiver** (`e *Event`), for exactly the reason Guide 6 gives: it needs to write the newly generated auto-increment ID back onto the caller's struct (`e.ID = id`), and a value receiver would only mutate a local copy that vanishes when the method returns. Contrast this with `Update` and `Delete`:

```go
func (event Event) Update() error { /* ... */ }
func (event Event) Delete() error { /* ... */ }
```

Both use **value receivers**, correctly — neither one needs to write anything back onto the struct, they only read fields to build a `WHERE id = ?` clause. `?` placeholders throughout (rather than string-concatenating values into the query) are the SQL-injection-safe parameterized-query pattern — the Go equivalent of using `cursor.execute(query, params)` in Python instead of an f-string, and just as non-negotiable for any code accepting user input.

`GetAllEvents` and `GetEventByID` show the read side of `database/sql`:

```go
func GetAllEvents() ([]Event, error) {
	rows, err := db.DB.Query("SELECT * FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.DateTime, &event.UserID)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
```

`Query` (plural rows) vs. `QueryRow` (single expected row, used in `GetEventByID`) is `database/sql`'s equivalent of `cursor.fetchall()` vs. `cursor.fetchone()`. `rows.Scan(&event.ID, &event.Name, ...)` takes a pointer per column, in the exact order the `SELECT` returns them — this is positional, not name-based, so if the `SELECT * ` column order and the `Scan` argument order ever drift out of sync (e.g., someone adds a column to the table without updating every `Scan` call), you get silently wrong data assigned to the wrong fields, not a compile error. `defer stmt.Close()` / `defer rows.Close()` releasing the resource right after acquiring it is the exact `defer`-for-cleanup pattern the notes' concurrency/pointers material establishes as the default reflex for anything with an OS-level resource attached (a file handle, a network connection, here a prepared statement or an open result cursor).

### 6.3 The `User` struct — and a real bug to find

```go
type User struct {
	ID       int64
	Email    string `binding:"required"`
	Password string `binding:"required"`
}

func (u User) Save() error {
	// ... hash password, INSERT, get LastInsertId() ...
	userId, err := result.LastInsertId()
	u.ID = userId
	return err
}
```

Compare this to `Event.Save()` above. `Event.Save` uses `(e *Event)` and correctly mutates the caller's ID. `User.Save` uses `(u User)` — a **value receiver** — and still tries `u.ID = userId`. This compiles fine and runs without error, but `u` is a local copy inside `Save`; the assignment is thrown away the instant `Save` returns, and the caller's `User.ID` stays `0` forever. It's a real, live bug, left in deliberately as a diagnostic: if you can spot why this line does nothing before reading this paragraph, Guide 6's pointer-vs-value-receiver material has fully landed. (It happens to be harmless here only because no caller currently reads `user.ID` after calling `Save` — see Section 15 for the fix.)

`ValidateCredentials`, by contrast, correctly uses a pointer receiver, because it *does* need the caller to see the fetched ID:

```go
func (u *User) ValidateCredentials() error {
	query := "SELECT id, password FROM users WHERE email = ?"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword)
	if err != nil {
		return errors.New("Credentials invalid")
	}

	if !utils.CheckPasswordHash(u.Password, retrievedPassword) {
		return errors.New("Credentials invalid")
	}
	return nil
}
```

Notice the deliberately vague error message: a real login failure could mean "no such email" (the `Scan` fails because `QueryRow` found zero rows) or "wrong password" (the `CheckPasswordHash` branch) — and both paths return the exact same `"Credentials invalid"` string. This is a genuine security practice, not sloppiness: telling an attacker *which* part failed ("no account with that email" vs. "wrong password") lets them enumerate valid emails one login attempt at a time. Collapsing both failure modes into one indistinguishable message is the standard mitigation.

---

## 7. Routing with Gin: `routes/routes.go`

```go
func RegisterRoutes(server *gin.Engine) {
	server.GET("/events", getEvents)
	server.GET("/events/:id", getEvent)

	authenticated := server.Group("/")
	authenticated.Use(middlewares.Authenticate)
	authenticated.POST("/events", createEvent)
	authenticated.PUT("/events/:id", updateEvent)
	authenticated.DELETE("/events/:id", deleteEvent)
	authenticated.POST("/events/:id/register", registerForEvent)
	authenticated.DELETE("/events/:id/register", cancelRegistration)

	server.POST("/signup", signup)
	server.POST("/login", login)
}
```

`:id` is a Gin **route parameter** — a path segment captured and made available inside the handler via `context.Param("id")`, the same mechanic as FastAPI's `@app.get("/events/{id}")` or Flask's `<id>`. The two `GET` routes are registered directly on `server` with no auth requirement, since browsing events shouldn't require being logged in.

`server.Group("/")` creates a **route group** — everything registered on `authenticated` instead of `server` shares whatever middleware is `.Use()`'d on the group. The `"/"` argument here is a path *prefix* for the group, and it happens to be the empty/root prefix — the group exists purely to attach `middlewares.Authenticate` to a specific subset of routes, not to change any URL. Every route added to `authenticated` runs `Authenticate` first (Section 11); if it calls `context.Abort...`, the handler function (`createEvent`, `updateEvent`, etc.) never runs at all. This is Gin's version of a FastAPI `Depends(get_current_user)` dependency, or a Django `LoginRequiredMixin` — the same idea (gate a set of routes behind an auth check) expressed as an explicit middleware chain rather than a decorator or mixin.

---

## 8. Route Handlers: Events, Signup, Login, Registration

Every handler in this project follows the same four-beat shape, which is worth internalizing since you'll write it dozens of times in any Gin/Go API:

1. Parse input (path param and/or JSON body).
2. On parse failure, respond `400` and `return` immediately.
3. Do the actual work (call into `models/`).
4. On a model/DB failure, respond `500` (or a more specific code) and `return`; otherwise respond success.

### 8.1 Listing and fetching events

```go
func getEvent(context *gin.Context) {
	eventId, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse event id."})
		return
	}

	event, err := models.GetEventByID(eventId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not fetch event."})
		return
	}

	context.JSON(http.StatusOK, event)
}
```

`context.Param("id")` always returns a `string` (path segments are text, full stop), so turning `"42"` into an `int64` needs an explicit `strconv.ParseInt` — the same "parse untrusted external input, then check the error" reflex Guide 5 establishes generally, just applied to a URL segment instead of a JSON body. Note this handler collapses "event not found" and "database error" into the same `500` — `sql.ErrNoRows` (what `QueryRow.Scan` returns when nothing matches) is treated identically to a real connection failure. A more precise version would check `errors.Is(err, sql.ErrNoRows)` and respond `404` instead (see Section 15).

`gin.H` is just `map[string]any` under a short alias — Gin's convenience type for "a JSON object I'm building inline," so `gin.H{"message": "..."}` marshals to `{"message": "..."}` the same way any `map[string]any` would via `encoding/json` (Guide 9, Section 8's territory, just going the *output* direction this time).

### 8.2 Signup and login

```go
func signup(context *gin.Context) {
	var user models.User
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}
	if err := user.Save(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save user."})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}
```

`context.ShouldBindJSON(&user)` is doing three things from Guide 9 in one call: reading the request body as a stream (`json.NewDecoder(r.Body)` under the hood), unmarshaling it into `user`, and running `go-playground/validator` against the `binding:"..."` tags — Gin's version of `decodeAndValidate` from Guide 9, Section 14, except pre-built into the framework instead of hand-rolled. A single `ShouldBindJSON` call replaces both the manual decode step and the manual `validate.Struct` call that guide walked through by hand.

`login` follows the same shape, then does one more thing signup doesn't:

```go
token, err := utils.GenerateToken(user.Email, user.ID)
// ...
context.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token})
```

`user.ID` here comes from `ValidateCredentials`'s pointer-receiver `Scan` (Section 6.3) — this is exactly why that method *needed* a pointer receiver and `User.Save` (arguably) doesn't: the ID fetched during login has to survive back out to the code that mints the JWT.

### 8.3 Creating and updating events: ownership checks

```go
func createEvent(context *gin.Context) {
	var event models.Event
	if err := context.ShouldBindJSON(&event); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	userId := context.GetInt64("userId")
	event.UserID = userId

	if err := event.Save(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create event. Try again later."})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Event created!", "event": event})
}
```

`context.GetInt64("userId")` reads a value stashed on the request context earlier in the chain — specifically, by the auth middleware (Section 11), which only runs at all because this route is registered on the `authenticated` group (Section 7). This is the load-bearing link between routing, middleware, and handler: **the handler trusts `userId` unconditionally**, with zero re-validation, because the middleware contract guarantees it only ever runs after a token has already been verified. Overwriting `event.UserID` with the authenticated user's ID (rather than trusting any `user_id` the client might have put in the request JSON) is what stops one user from creating events "owned" by someone else.

`updateEvent` and `deleteEvent` add an explicit **ownership check** the create path doesn't need:

```go
event, err := models.GetEventByID(eventId)
// ...
if event.UserID != userId {
	context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized to update event."})
	return
}
```

Authentication (the middleware: "is this a valid token, and whose?") and **authorization** (this check: "is this specific user allowed to touch this specific resource?") are two genuinely different questions, and this is where the project draws the line between them — the middleware answers the first, individual handlers answer the second, on a per-resource basis, since only the handler knows which resource is even being requested.

---

## 9. Password Hashing: `utils/hash.go`

```go
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
```

`bcrypt` is a deliberately **slow** hashing algorithm — the `14` is the cost factor (work factor), controlling how many rounds of internal hashing run; higher costs mean more CPU time per hash, which is exactly the point: it makes brute-forcing a stolen password database computationally expensive, at the cost of a few hundred milliseconds per login. This is the same algorithm and the same reasoning Python's `bcrypt`/`passlib` libraries use — never hash passwords with a fast general-purpose hash like SHA-256 or MD5, which are *designed* to be fast and are correspondingly terrible at resisting brute force. `CompareHashAndPassword` deliberately never returns the hash itself for comparison — you always ask bcrypt "does this plaintext match this stored hash," never extract and compare hashes manually, since bcrypt hashes embed a random salt that makes two hashes of the same password look completely different by design.

---

## 10. JWT Authentication: `utils/jwt.go`

```go
const secretKey = "supersecret"

func GenerateToken(email string, userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  email,
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 2).Unix(),
	})
	return token.SignedString([]byte(secretKey))
}
```

A JWT is three base64-encoded, dot-separated segments — header, payload (**claims**), signature — and `jwt.MapClaims` is just `map[string]any` under an alias, the same "untyped JSON bag" pattern as `gin.H`. `HS256` is HMAC-SHA256: a **symmetric** signing scheme, meaning the exact same `secretKey` both signs new tokens and verifies incoming ones (as opposed to an asymmetric scheme like RS256, where a private key signs and a separate public key verifies — overkill for a single-server project like this one, but the standard choice once multiple independent services need to verify tokens without all sharing one secret). `exp` (expiration, a Unix timestamp) is a **registered claim name** the JWT spec gives special meaning to — the library automatically rejects an expired token during verification, without you writing any expiration-checking logic by hand.

```go
func VerifyToken(token string) (int64, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("Unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	// ...
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	// ...
	userId := int64(claims["userId"].(float64))
	return userId, nil
}
```

Two details worth slowing down on:

- The callback checking `token.Method.(*jwt.SigningMethodHMAC)` (a **type assertion**, Guide 4) before trusting the token is a defense against a real, historical JWT vulnerability: an attacker-supplied token can claim *any* algorithm in its own header, including `"none"` or an asymmetric algorithm your server would verify with the wrong key material. Never trust the algorithm a token claims for itself without checking it's one your server actually expects — this callback exists specifically to enforce that.
- `int64(claims["userId"].(float64))` is a direct, real-world hit of **Guide 9, Section 8's `float64` trap**. JWT claims are transmitted as JSON internally, and this library decodes them into a `map[string]any` — so `userId`, even though it was written in as an `int64` during `GenerateToken`, comes back out of `MapClaims` as a `float64`, and asserting straight to `int64` here would panic. The two-step `claims["userId"].(float64)` then `int64(...)` is exactly the fix that guide's Section 8 walks through in the abstract — this is what it looks like showing up in a real dependency's API rather than a contrived example.

The commented-out line `// email := claims["email"].(string)` is a leftover — the email claim is embedded in every token but nothing in this codebase currently reads it back out.

---

## 11. The Auth Middleware: `middlewares/auth.go`

```go
func Authenticate(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")

	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	userId, err := utils.VerifyToken(token)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	context.Set("userId", userId)
	context.Next()
}
```

This is a Gin **middleware**: a function taking `*gin.Context` that runs *before* the actual route handler, with the power to either let the request continue or stop it dead. The two calls that make this work:

- **`context.AbortWithStatusJSON`** writes the response and marks the context as aborted — Gin will not call any further middleware or the route handler after this. Using `Abort...` instead of a plain `JSON` call here matters: without it, execution would fall through to `context.Next()` regardless, defeating the entire point of the check.
- **`context.Set("userId", userId)`** stores a value on the request-scoped context, retrievable later via `context.GetInt64("userId")` exactly as seen in `createEvent` (Section 8.3). This is Gin's per-request key-value store — a lightweight version of what Guide 7's `context.Context` `WithValue` does in the standard library, scoped to a single request's lifetime and cleaned up automatically once the response is written.
- **`context.Next()`** explicitly hands control to whatever comes next in the chain (the next middleware, or the final route handler). Forgetting this call anywhere in a middleware silently stalls every request that passes through it — worth remembering as the single easiest way to accidentally hang a Gin server that has no middleware bug reported anywhere in its logs.

Note the middleware reads the raw `Authorization` header value directly as the token (`token := context.Request.Header.Get("Authorization")`), with no `"Bearer "` prefix stripped — this project expects clients to send the bare JWT as the header value, not the more common `Authorization: Bearer <token>` convention. Check the `api-test/create-event.http` file (Section 13) — it sends the raw token with no `Bearer` prefix, consistent with what this middleware expects.

---

## 12. Full Request Walkthrough: Creating an Event

Tracing one request through every layer, in order, ties the whole project together:

1. A client sends `POST /events` with header `Authorization: <jwt>` and a JSON body (`name`, `description`, `location`, `dateTime`).
2. Gin matches the route (`routes.go`), sees it belongs to the `authenticated` group, and runs `middlewares.Authenticate` first.
3. `Authenticate` reads the header, calls `utils.VerifyToken`, which parses the JWT, checks its signing method, verifies its signature against `secretKey`, checks it hasn't expired, and extracts `userId` (working around the `float64` trap along the way). On success, `context.Set("userId", userId)` and `context.Next()`.
4. Gin now calls `createEvent` (`routes/events.go`). `context.ShouldBindJSON(&event)` reads the body stream, unmarshals it into an `Event`, and runs `binding:"required"` validation on `Name`, `Description`, `Location`, `DateTime` — any missing field aborts here with a `400`.
5. The handler reads `userId` back off the context and assigns it to `event.UserID` — this is the only place the event's ownership is decided, and it comes from the verified token, never from client-supplied JSON.
6. `event.Save()` (`models/event.go`) prepares a parameterized `INSERT`, executes it against `db.DB` (the shared connection pool from `db/db.go`), and — because `Save` has a pointer receiver — writes the new auto-increment ID back onto `event`.
7. The handler responds `201 Created` with the full event (now including its real `ID`) as JSON.

Every arrow in that chain is a function call you can `Ctrl`-click to in your editor — there's no framework-level reflection deciding which handler or middleware runs based on struct field names or decorators, the way Django's URL resolver or FastAPI's dependency graph works. `routes.go`'s explicit `.Use()` and `.POST(path, handler)` calls *are* the entire routing table.

---

## 13. Manual Testing: the `api-test/*.http` Files

The `api-test/` folder holds plain-text HTTP request files, meant for the "REST Client"-style extension in VS Code (or any editor with equivalent support) — each `.http` file is a runnable request you can fire with one click, no `curl` typing required:

```http
POST http://localhost:8080/signup
content-type: application/json

{
  "email": "test2@example.com",
  "password": "test"
}
```

```http
POST http://localhost:8080/login
content-type: application/json

{
  "email": "test2@example.com",
  "password": "test"
}
```

```http
POST http://localhost:8080/events
content-type: application/json
authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "name": "Test event",
  "description": "Test event!!!",
  "location": "A test location",
  "dateTime": "2025-01-01T15:30:00.000Z"
}
```

The intended workflow: run `create-user.http` (signup) once, then `login.http` to get a fresh token back in the response body, paste that token into the `authorization:` line of `create-event.http` (and the other authenticated `.http` files), then run those. The tokens hardcoded in these files are already expired (2 hours from whenever they were originally generated) — you'll need to log in again and paste a fresh one before they'll work. `dateTime` is sent as an RFC 3339 string (`"2025-01-01T15:30:00.000Z"`) — exactly the format Guide 9, Section 12 covers as `time.Time`'s default JSON representation, which is why `Event.DateTime time.Time` needs no custom marshaling code at all to accept it.

---

## 14. Running the Project

```bash
cd hands-on/11-REST-API
go run main.go
```

This compiles and runs `main.go` directly — no separate build step needed for local development (`go build` produces a standalone binary if you want one). On first run, `db.InitDB()` creates `api.db` in the current directory and the three tables inside it; on every subsequent run, `CREATE TABLE IF NOT EXISTS` is a no-op against the existing file, so your data persists across restarts. Delete `api.db` at any point to reset to a clean slate. The server listens on `http://localhost:8080` — use the `.http` files (Section 13) or any HTTP client (`curl`, Postman, Insomnia) to exercise it once it's running.

---

## 15. Rough Edges: Bugs and Exercises

This project is a capstone, not production code, and a few real issues are worth finding and fixing yourself as a final exercise in reading Go critically rather than trustingly:

1. **`User.Save`'s lost ID (Section 6.3).** Change the receiver from `(u User)` to `(u *User)` and update every call site (`user.Save()` still works unchanged — Go automatically takes the address of an addressable value when calling a pointer-receiver method). Confirm with a quick test that `user.ID` is non-zero immediately after `Save()` returns.
2. **The swallowed `ParseInt` error in `cancelRegistration`** (`routes/register.go`): `eventId, err := strconv.ParseInt(...)` is called, but `err` is never checked before being overwritten by the following `event.CancelRegistration(userId)` call's return value. Add the missing `if err != nil { ... }` block (compare against `registerForEvent` in the same file, which checks it correctly), and consider what request would currently slip through silently without the fix.
3. **`404` vs. `500` for "not found."** Every `GetEventByID` failure currently returns `500 Internal Server Error`, even when the real cause is simply "no event with that ID" (`sql.ErrNoRows`). Use `errors.Is(err, sql.ErrNoRows)` in the handlers to return `404 Not Found` specifically for that case, and reserve `500` for genuine database failures.
4. **Hardcoded secret key.** `secretKey = "supersecret"` in `utils/jwt.go` is checked directly into source control — anyone reading this repository can forge valid tokens for any user. Move it to an environment variable (`os.Getenv("JWT_SECRET")`), loaded once at startup, and add `api.db` and any `.env` file to `.gitignore` if they aren't already.
5. **No `Bearer` prefix convention.** Section 11 noted the middleware expects the raw token with no `Authorization: Bearer <token>` prefix. Update `Authenticate` to strip a `"Bearer "` prefix if present (falling back to the raw value otherwise), which is the more widely-expected convention for HTTP clients and libraries.
6. **No table-driven tests exist yet.** Guide 8 covers exactly the tooling this project has none of. Pick one pure-ish function — `utils.HashPassword`/`CheckPasswordHash` round-tripping, or `utils.GenerateToken`/`VerifyToken` round-tripping — and write a table-driven test file for it as practice applying that guide to real project code instead of a toy example.

---

## 16. Key Takeaways

- This project is the practical synthesis of every prior guide: **structs and methods** (Guide 3) model `Event`/`User`; **interfaces** (Guide 4) make `database/sql` driver-agnostic; **explicit error returns** (Guide 5) gate every handler step; **pointer vs. value receivers** (Guide 6) decide whether a `.Save()` call's generated ID survives back to the caller — and getting it wrong (`User.Save`) is a real, silent bug, not a hypothetical one; **JSON marshaling and the `float64` trap** (Guide 9) show up unmodified inside a third-party JWT library's claims map.
- **Gin** layers routing, route parameters, middleware groups, and request binding/validation (`ShouldBindJSON` + `binding:"..."` tags) on top of `net/http` — structurally the same job Flask/FastAPI's routing and Pydantic-style validation do, just wired together explicitly through `.Use()` and struct tags rather than decorators or dependency injection.
- **Authentication** (verifying *who* is asking — the JWT middleware) and **authorization** (deciding *what* they're allowed to do — the per-handler ownership checks in `updateEvent`/`deleteEvent`) are separate concerns in this codebase, deliberately: the middleware only ever answers the first question, and only a handler that knows which specific resource is being touched can answer the second.
- **`database/sql` without an ORM** means every query is a literal SQL string with `?` placeholders (never string-concatenated — that's SQL injection) and every result row is manually `Scan`'d into struct fields, in positional column order — more typing than an ORM, but nothing about how a query executes is hidden from you.
- **Passwords are bcrypt-hashed, never stored or compared as plaintext**, and login failures are deliberately vague ("Credentials invalid" for both "no such user" and "wrong password") to avoid leaking which emails have accounts.
- **JWTs carry their claims as a JSON map**, are signed with a symmetric secret (`HS256`) that must be checked against the token's own claimed algorithm before trusting it, and carry their own expiration (`exp`) that the verifying library checks automatically.
- Reading a finished project critically — not just running it — is its own skill: Section 15's five issues are all things that compile, run, and mostly work, and none of them would show up without either testing the right edge case or reading every line with the same skepticism you'd apply to someone else's pull request.
