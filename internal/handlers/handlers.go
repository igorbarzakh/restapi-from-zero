package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"restapi-tasks/internal/database"
	"restapi-tasks/internal/models"
	"strconv"
	"strings"
)

type TaskStore interface {
	GetAll() ([]models.Task, error)
	GetByID(int) (*models.Task, error)
	Create(models.CreateTaskInput) (*models.Task, error)
	Update(int, models.UpdateTaskInput) (*models.Task, error)
	Delete(int) error
}

type Handlers struct {
	store TaskStore
}

func NewHandler(store TaskStore) *Handlers {
	return &Handlers{store: store}
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}

func parseTaskID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func (h *Handlers) GetAllTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := h.store.GetAll()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

func (h *Handlers) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}
	task, err := h.store.GetByID(id)

	if err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			respondWithError(w, http.StatusNotFound, "Task not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input models.CreateTaskInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	task, err := h.store.Create(input)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, task)
}

func (h *Handlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var input models.UpdateTaskInput

	err = json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	task, err := h.store.Update(id, input)

	if err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	err = h.store.Delete(id)
	if err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
