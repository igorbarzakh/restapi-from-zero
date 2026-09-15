package main

import (
	"log"
	"net/http"
	"os"
	"restapi-tasks/internal/database"
	"restapi-tasks/internal/handlers"
)

func main() {
	dataBaseURL := os.Getenv("DATABASE_URL")
	serverPort := os.Getenv("SERVER_PORT")

	log.Printf("Starting server on port %s", serverPort)

	db, err := database.Connect(dataBaseURL)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	log.Printf("Database connection established")

	taskStore := database.NewTaskStore(db)

	handler := handlers.NewHandler(taskStore)

	mux := http.NewServeMux()

	mux.Handle("/tasks", methodHandler(handler.GetAllTasks, "GET"))
	mux.Handle("/tasks/create", methodHandler(handler.CreateTask, "POST"))

	mux.HandleFunc("/tasks/{id}", taskIDHandler(handler))

	loggedMux := loggingMiddleware(mux)

	// TODO: Implement CORS

	serverAddress := ":" + serverPort

	log.Printf("Listening on port %s", serverPort)

	err = http.ListenAndServe(serverAddress, loggedMux)

	if err != nil {
		log.Fatal(err)
	}

}

func methodHandler(handlerFunc http.HandlerFunc, allowMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowMethod {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		handlerFunc(w, r)
	}
}

func taskIDHandler(handlerFunc *handlers.Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlerFunc.GetTaskByID(w, r)
		case http.MethodPut:
			handlerFunc.UpdateTask(w, r)
		case http.MethodDelete:
			handlerFunc.DeleteTask(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
