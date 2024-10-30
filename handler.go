package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/meenaravichandran1/runner-logger/logger"
	"io"
	"net/http"
	"sync"
)

type Handler struct {
	port  string
	tasks map[string]chan bool
	lock  sync.Mutex
}

func (h *Handler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling tasks in daemonset-simple-task handler")
	switch r.Method {
	case http.MethodPost:
		h.Assign(w, r)
	case http.MethodDelete:
		h.Remove(w, r)
	case http.MethodGet:
		h.Get(w, r)
	default:
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	var tasks Tasks
	if err := parseRequest(r, &tasks); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	for _, task := range tasks.Tasks {
		var params Params
		if err := parseParams(&task.EncodedParams, &params); err != nil {
			sendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		quit := make(chan bool)
		go startTask(task.ID, params, quit)

		h.lock.Lock()
		h.tasks[task.ID] = quit
		h.lock.Unlock()
	}

	sendSuccessResponse(w, Response{TasksMetadata: h.getTasksMetadata()})
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	taskIds, ok := r.URL.Query()["taskIds"]
	if !ok || len(taskIds) < 1 {
		sendErrorResponse(w, http.StatusBadRequest, "task IDs are required")
		return
	}

	h.lock.Lock()

	for _, taskId := range taskIds {
		if quit, exists := h.tasks[taskId]; exists {
			go func(quit chan bool) { quit <- true }(quit)
			delete(h.tasks, taskId)
		} else {
			sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("the task with ID %s does not exist!", taskId))
			return
		}
	}

	h.lock.Unlock()

	sendSuccessResponse(w, Response{TasksMetadata: h.getTasksMetadata()})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	sendSuccessResponse(w, Response{TasksMetadata: h.getTasksMetadata()})
}

func (h *Handler) getTasksMetadata() TasksMetadata {
	h.lock.Lock()
	defer h.lock.Unlock()

	var tasksMetadata TasksMetadata
	for id := range h.tasks {
		tasksMetadata = append(tasksMetadata, TaskMetadata{ID: id})
	}
	return tasksMetadata
}

func parseRequest(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body")
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("invalid payload")
	}
	return nil
}

func parseParams(e *EncodedParams, p *Params) error {
	// decode base64 data
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(e.Base64Data)))
	n, err := base64.StdEncoding.Decode(decoded, e.Base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode EncodedParams.Base64Data with base64: %w", err)
	}
	decoded = decoded[:n]
	// unmarshall decoded data into `Params` type
	if err := json.Unmarshal(decoded, p); err != nil {
		return fmt.Errorf("decoded value of EncodedParams.Base64Data is not valid Params type: %w", err)
	}
	return nil
}

func sendSuccessResponse(w http.ResponseWriter, response interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, status int, errMsg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": errMsg,
	})
}
