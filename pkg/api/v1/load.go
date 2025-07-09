package v1

import (
	"encoding/json"
	"net/http"
	"waitingroom/internal/services/load"
)

type LoadHandler interface {
	LoadStatusHandler(w http.ResponseWriter, r *http.Request)
}

type LoadHTTPHandler struct {
	loadService load.Service
}

func NewLoadHandler() LoadHandler {
	return &LoadHTTPHandler{
		loadService: load.NewService(),
	}
}

func (h *LoadHTTPHandler) LoadStatusHandler(w http.ResponseWriter, r *http.Request) {
	loadModel := h.loadService.GetStatus(r.Context())

	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"is_reached": loadModel.ReachedCapacity(),
		"count":      loadModel.Count,
	}
	json.NewEncoder(w).Encode(response)

	return
}
