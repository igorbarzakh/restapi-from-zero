# Tasks REST API

A small Go REST API for creating, listing, updating, and deleting tasks stored in PostgreSQL.

## Stack

- Go 1.27.0 (as specified in `go.mod`)
- Standard library `net/http` router and JSON encoding
- PostgreSQL 15
- `sqlx` and the `lib/pq` PostgreSQL driver
- Docker Compose for the database

## Project structure

```text
cmd/api/main.go                 Application startup, routes, and request logging
internal/handlers/handlers.go  HTTP handlers and input validation
internal/database/database.go Database connection and connection pool
internal/database/tasks.go    Task queries
internal/models/task.go       Task model and request types
sql/init.sql                  Initial schema and sample tasks
docker-compose.yml            PostgreSQL service and persistent volume
```

## Local setup

Run the following commands from the project root. You need Go and a running Docker installation with Docker Compose.

### 1. Configure the environment

Create a `.env` file with the following example values. If the file already exists, keep your existing configuration and ensure the connection URL matches your database credentials.

```dotenv
POSTGRES_DB=tasks
POSTGRES_USER=tasks
POSTGRES_PASSWORD=local_dev_password
DATABASE_URL='postgres://tasks:local_dev_password@localhost:5432/tasks?sslmode=disable'
SERVER_PORT=8080
```

These credentials are for local development. `.env` is excluded from Git.

Docker Compose reads `.env` automatically. The Go application reads environment variables and does not load `.env` itself.

### 2. Start PostgreSQL

```sh
docker compose up -d postgres
```

Check that the database is ready before starting the API:

```sh
docker compose exec postgres sh -c 'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
```

On the first startup with an empty database volume, `sql/init.sql` creates the `tasks` table and inserts three sample tasks. It is not rerun on subsequent starts with the same volume. The script contains `DROP TABLE IF EXISTS tasks`, so manually rerunning it replaces existing task data.

### 3. Start the API

Load the local `.env` file into your shell and run the application (Bash or Zsh):

```sh
set -a
source .env
set +a
go run ./cmd/api
```

With the example configuration, the API is available at `http://localhost:8080`. Docker Compose runs only PostgreSQL; the API runs on your machine.

## API

| Method | Path | Description | Success status |
| --- | --- | --- | --- |
| GET | `/tasks` | List tasks, newest first | 200 |
| GET | `/tasks/{id}` | Get a task | 200 |
| POST | `/tasks/create` | Create a task | 201 |
| PUT | `/tasks/{id}` | Update supplied task fields | 200 |
| DELETE | `/tasks/{id}` | Delete a task | 200 |

Task responses contain `id`, `title`, `description`, `completed`, `created_at`, and `updated_at`.

### List tasks

```sh
curl http://localhost:8080/tasks
```

### Get a task

```sh
curl http://localhost:8080/tasks/1
```

### Create a task

```sh
curl -X POST http://localhost:8080/tasks/create \
  -H 'Content-Type: application/json' \
  -d '{"title":"Learn Go","description":"Build a REST API","completed":false}'
```

`title` is required and must not be blank. `description` and `completed` default to an empty string and `false` when omitted. The database limits titles to 255 characters.

### Update a task

```sh
curl -X PUT http://localhost:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"completed":true}'
```

Although the endpoint uses `PUT`, it performs a partial update: omitted fields remain unchanged. You can explicitly send `false` for `completed` or an empty string for `description`. If supplied, `title` must not be blank. JSON `null` is treated as an omitted field.

### Delete a task

```sh
curl -X DELETE http://localhost:8080/tasks/1
```

Successful deletion returns:

```json
{"result":"success"}
```

### Errors

Handler errors use this JSON format:

```json
{"error":"Task not found"}
```

- `400`: invalid ID, malformed request payload, or blank title.
- `404`: task not found.
- `405`: unsupported HTTP method (plain-text response).
- `500`: database or other internal error.

## Checks

```sh
go test ./...
go vet ./...
```

There are currently no automated test files; `go test` checks that the packages compile.

## Stop locally

Stop the API with `Ctrl+C`, then stop the database:

```sh
docker compose down
```

The named database volume is preserved.
