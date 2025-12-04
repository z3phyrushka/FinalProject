package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"Final-project/pkg/db"
)

func writeJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t db.Task

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&t); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(t.ID) == "" {
		writeJSON(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	if len(t.Title) == 0 {
		writeJSON(w, map[string]string{"error": "title is required"}, http.StatusBadRequest)
		return
	}

	nowTime := time.Now().UTC()
	now := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.UTC)

	if err := checkDate(&t, now); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if err := db.UpdateTask(&t); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{}, http.StatusOK)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t db.Task
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&t); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(t.Title) == "" {
		writeJSON(w, map[string]string{"error": "title is required"}, http.StatusBadRequest)
		return
	}

	nowTime := time.Now().UTC()
	now := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.UTC)

	if err := checkDate(&t, now); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&t)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{"id": id}, http.StatusOK)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "id is required"}, http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{}, http.StatusOK)
}

func checkDate(t *db.Task, now time.Time) error {
	if t.Date == "" {
		t.Date = now.Format(dateFormat)
	}

	parsed, err := time.Parse(dateFormat, t.Date)
	if err != nil {
		return errors.New("invalid date format")
	}
	parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)

	var next string
	if strings.TrimSpace(t.Repeat) != "" {
		next, err = NextDate(now, t.Date, t.Repeat)
		if err != nil {
			return fmt.Errorf("invalid repeat: %w", err)
		}
	}

	if afterNow(now, parsed) {
		if strings.TrimSpace(t.Repeat) == "" {
			t.Date = now.Format(dateFormat)
		} else {
			t.Date = next
		}
	}

	return nil
}
