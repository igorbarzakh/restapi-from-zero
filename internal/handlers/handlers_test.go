package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"restapi-tasks/internal/database"
	"restapi-tasks/internal/models"
)

// Unset callbacks make unexpected storage calls fail instead of silently succeeding.
type stubStore struct {
	getAll  func() ([]models.Task, error)
	getByID func(int) (*models.Task, error)
	create  func(models.CreateTaskInput) (*models.Task, error)
	update  func(int, models.UpdateTaskInput) (*models.Task, error)
	delete  func(int) error
}

func (s stubStore) GetAll() ([]models.Task, error)                         { return s.getAll() }
func (s stubStore) GetByID(id int) (*models.Task, error)                   { return s.getByID(id) }
func (s stubStore) Create(in models.CreateTaskInput) (*models.Task, error) { return s.create(in) }
func (s stubStore) Update(id int, in models.UpdateTaskInput) (*models.Task, error) {
	return s.update(id, in)
}
func (s stubStore) Delete(id int) error { return s.delete(id) }

func request(t *testing.T, handler http.HandlerFunc, method, id, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, "/tasks/"+id, strings.NewReader(body))
	r.SetPathValue("id", id)
	w := httptest.NewRecorder()
	handler(w, r)
	if w.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, status, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	return w
}

func TestInvalidRequests(t *testing.T) {
	h := NewHandler(stubStore{})
	cases := []struct {
		name, method, id, body, message string
		handler                         http.HandlerFunc
	}{
		{"get invalid ID", "GET", "abc", "", "Invalid task ID", h.GetTaskByID},
		{"update invalid ID", "PUT", "abc", `{}`, "Invalid task ID", h.UpdateTask},
		{"delete invalid ID", "DELETE", "abc", "", "Invalid task ID", h.DeleteTask},
		{"create malformed JSON", "POST", "", `{`, "Invalid request payload", h.CreateTask},
		{"create empty body", "POST", "", "", "Invalid request payload", h.CreateTask},
		{"create missing title", "POST", "", `{}`, "Title is required", h.CreateTask},
		{"create blank title", "POST", "", `{"title":" \t "}`, "Title is required", h.CreateTask},
		{"create wrong field type", "POST", "", `{"title":123}`, "Invalid request payload", h.CreateTask},
		{"update malformed JSON", "PUT", "1", `{`, "Invalid request payload", h.UpdateTask},
		{"update blank title", "PUT", "1", `{"title":" "}`, "Title is required", h.UpdateTask},
		{"update wrong field type", "PUT", "1", `{"completed":"yes"}`, "Invalid request payload", h.UpdateTask},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := request(t, tc.handler, tc.method, tc.id, tc.body, 400)
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body["error"] != tc.message {
				t.Fatalf("error = %q, want %q", body["error"], tc.message)
			}
		})
	}
}

func TestStorageErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"not found", fmt.Errorf("lookup: %w", database.ErrTaskNotFound), 404},
		{"database failure", fmt.Errorf("database unavailable"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := stubStore{
				getByID: func(id int) (*models.Task, error) {
					if id != 42 {
						t.Errorf("id = %d", id)
					}
					return nil, tc.err
				},
				update: func(id int, _ models.UpdateTaskInput) (*models.Task, error) {
					if id != 42 {
						t.Errorf("id = %d", id)
					}
					return nil, tc.err
				},
				delete: func(id int) error {
					if id != 42 {
						t.Errorf("id = %d", id)
					}
					return tc.err
				},
			}
			h := NewHandler(store)
			for _, endpoint := range []struct {
				method  string
				handler http.HandlerFunc
			}{{"GET", h.GetTaskByID}, {"PUT", h.UpdateTask}, {"DELETE", h.DeleteTask}} {
				t.Run(endpoint.method, func(t *testing.T) {
					w := request(t, endpoint.handler, endpoint.method, "42", `{}`, tc.status)
					var body map[string]string
					if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
						t.Fatal(err)
					}
					if body["error"] == "" {
						t.Fatal("missing error message")
					}
				})
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	for _, body := range []string{`{"title":"Learn Go"}`, `{"title":"Learn Go","description":"Read docs","completed":true}`} {
		t.Run(body, func(t *testing.T) {
			var want models.CreateTaskInput
			if err := json.Unmarshal([]byte(body), &want); err != nil {
				t.Fatal(err)
			}
			calls := 0
			task := models.Task{ID: 42, Title: want.Title, Description: want.Description, Completed: want.Completed}
			h := NewHandler(stubStore{create: func(in models.CreateTaskInput) (*models.Task, error) {
				calls++
				if in != want {
					t.Fatalf("input = %+v, want %+v", in, want)
				}
				return &task, nil
			}})
			w := request(t, h.CreateTask, "POST", "", body, 201)
			var got models.Task
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got != task || calls != 1 {
				t.Fatalf("task = %+v, calls = %d", got, calls)
			}
		})
	}
}

func TestUpdateTaskPartialInput(t *testing.T) {
	empty := ""
	completed := false
	title := "New title"
	for _, tc := range []struct {
		name, body string
		want       models.UpdateTaskInput
	}{
		{"omitted fields", `{}`, models.UpdateTaskInput{}},
		{"explicit false", `{"completed":false}`, models.UpdateTaskInput{Completed: &completed}},
		{"empty description", `{"description":""}`, models.UpdateTaskInput{Description: &empty}},
		{"title only", `{"title":"New title"}`, models.UpdateTaskInput{Title: &title}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			task := models.Task{ID: 42, Title: "New title"}
			h := NewHandler(stubStore{update: func(id int, in models.UpdateTaskInput) (*models.Task, error) {
				calls++
				if id != 42 || !reflect.DeepEqual(in, tc.want) {
					t.Fatalf("id = %d, input = %+v, want %+v", id, in, tc.want)
				}
				return &task, nil
			}})
			w := request(t, h.UpdateTask, "PUT", "42", tc.body, 200)
			var got models.Task
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got != task || calls != 1 {
				t.Fatalf("task = %+v, calls = %d", got, calls)
			}
		})
	}
}

func TestReadAndDeleteTasks(t *testing.T) {
	task := models.Task{ID: 42, Title: "Learn Go"}
	t.Run("list", func(t *testing.T) {
		h := NewHandler(stubStore{getAll: func() ([]models.Task, error) { return []models.Task{task}, nil }})
		w := request(t, h.GetAllTasks, "GET", "", "", 200)
		var got []models.Task
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, []models.Task{task}) {
			t.Fatalf("tasks = %+v", got)
		}
	})
	t.Run("get", func(t *testing.T) {
		h := NewHandler(stubStore{getByID: func(id int) (*models.Task, error) {
			if id != 42 {
				t.Fatalf("id = %d", id)
			}
			return &task, nil
		}})
		w := request(t, h.GetTaskByID, "GET", "42", "", 200)
		var got models.Task
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got != task {
			t.Fatalf("task = %+v", got)
		}
	})
	t.Run("delete", func(t *testing.T) {
		calls := 0
		h := NewHandler(stubStore{delete: func(id int) error {
			calls++
			if id != 42 {
				t.Fatalf("id = %d", id)
			}
			return nil
		}})
		w := request(t, h.DeleteTask, "DELETE", "42", "", 200)
		var got map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got["result"] != "success" || calls != 1 {
			t.Fatalf("body = %v, calls = %d", got, calls)
		}
	})
}

func TestListAndCreateStorageFailure(t *testing.T) {
	failure := fmt.Errorf("database unavailable")
	h := NewHandler(stubStore{
		getAll: func() ([]models.Task, error) { return nil, failure },
		create: func(models.CreateTaskInput) (*models.Task, error) { return nil, failure },
	})
	request(t, h.GetAllTasks, "GET", "", "", 500)
	request(t, h.CreateTask, "POST", "", `{"title":"Learn Go"}`, 500)
}
