package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"restapi-tasks/internal/handlers"
)

func TestMethodHandler(t *testing.T) {
	for _, tc := range []struct {
		name, method, allowed string
		status, calls         int
	}{
		{"allowed GET", "GET", "GET", 204, 1},
		{"allowed POST", "POST", "POST", 204, 1},
		{"reject GET on create", "GET", "POST", 405, 0},
		{"reject POST on list", "POST", "GET", 405, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			handler := methodHandler(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(204) }, tc.allowed)
			w := httptest.NewRecorder()
			handler(w, httptest.NewRequest(tc.method, "/tasks", nil))
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			if calls != tc.calls {
				t.Fatalf("handler called %d times, want %d", calls, tc.calls)
			}
		})
	}
}

func TestTaskIDHandlerDispatch(t *testing.T) {
	// Invalid IDs stop each supported handler before it needs a database.
	h := handlers.NewHandler(nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks/{id}", taskIDHandler(h))
	for _, tc := range []struct {
		method string
		status int
	}{
		{"GET", 400}, {"PUT", 400}, {"DELETE", 400}, {"POST", 405}, {"PATCH", 405},
	} {
		t.Run(tc.method, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(tc.method, "/tasks/invalid", nil))
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
		})
	}
}
