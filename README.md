# Tasks REST API

A small Go REST API for creating, listing, updating, and deleting tasks stored in PostgreSQL.

## Stack

- Go 1.27.0 (as specified in `go.mod`)
- Standard library `net/http` router and JSON encoding
- PostgreSQL 15
- `sqlx` and the `lib/pq` PostgreSQL driver
- Docker Compose for the API and database

## Project structure

```text
cmd/api/main.go                 Application startup, routes, and request logging
internal/handlers/handlers.go  HTTP handlers and input validation
internal/database/database.go Database connection and connection pool
internal/database/tasks.go    Task queries
internal/models/task.go       Task model and request types
sql/init.sql                  Initial schema and sample tasks
docker-compose.yml            API, PostgreSQL, and persistent volume
Dockerfile                    Multi-stage API image build
Makefile                      Development commands
```

## Quick start

Install Docker with Docker Compose and Make, then run from the project root:

```sh
make up
```

This builds the API image, starts PostgreSQL, waits for the database health check, and starts the API. With the default settings, the API is available at `http://localhost:8080`. Go does not need to be installed on your machine for this workflow.

Without Make, use the equivalent command:

```sh
docker compose up --build
```

The first build requires internet access to download images and Go dependencies. Run `make up` again after changing Go code to rebuild the API image; there is no automatic reload.

## Development commands

| Command | Description |
| --- | --- |
| `make up` | Build and start the API and database with logs in the terminal |
| `make down` | Stop and remove the containers, preserving database data |
| `make logs` | Follow API container logs from another terminal |
| `make test` | Run `go test ./...` using locally installed Go |
| `make run` | Run the API locally using Go; PostgreSQL must already be running |

For background startup, use `docker compose up --build -d`.

## Configuration

No `.env` file is required for a fresh local setup. Docker Compose uses these defaults, which you can override in a `.env` file in the project root:

```dotenv
POSTGRES_DB=tasks
POSTGRES_USER=tasks
POSTGRES_PASSWORD=local_dev_password
SERVER_PORT=8080
```

These credentials are for local development. `.env` is excluded from Git and the Docker build context. Keep your existing credentials if you already have a database volume: changing these variables does not change users or passwords in an initialized database.

The containerized API connects to `postgres:5432`. Compose supplies the database name, user, and password through `PGDATABASE`, `PGUSER`, and `PGPASSWORD`, and sets `DATABASE_URL` to the internal database address. A `DATABASE_URL` in your local `.env` is used only by `make run` and does not override the container connection.

PostgreSQL is exposed on port `5432`; the API is exposed on `SERVER_PORT` (default `8080`).

On the first startup with an empty database volume, `sql/init.sql` creates the `tasks` table and inserts three sample tasks. It is not rerun on subsequent starts with the same volume. The script contains `DROP TABLE IF EXISTS tasks`, so manually rerunning it replaces existing task data.

## Run the API without Docker

For local Go development, install the Go version specified in `go.mod`. Start only PostgreSQL, then run the API:

```sh
docker compose up -d --wait postgres
make run
```

`make run` loads `.env` when present and provides the same local defaults as Compose. The file must use shell-compatible assignments. To connect to a different database, set `DATABASE_URL`, for example:

```dotenv
DATABASE_URL='postgres://tasks:local_dev_password@localhost:5432/tasks?sslmode=disable'
```

Stop the containerized API before starting a local API on the same port (`docker compose stop api`). Stop the local API with `Ctrl+C`.

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
make test
go vet ./...
```

There are currently no automated test files; `go test` checks that the packages compile.

## Stop locally

Stop the local API with `Ctrl+C` if it is running, then stop the Compose services:

```sh
make down
```

The named database volume is preserved.
