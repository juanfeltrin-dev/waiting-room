package v1

import (
	"encoding/json"
	"net/http"
	"waitingroom/internal/infra/jwt"
	"waitingroom/internal/schemas/request"
	"waitingroom/internal/services/queue"
)

type QueueHandler interface {
	GetPositionHandler(w http.ResponseWriter, r *http.Request)
	EnterHandler(w http.ResponseWriter, r *http.Request)
	ExitHandler(w http.ResponseWriter, r *http.Request)
	GetAverageQueueTimeHandler(w http.ResponseWriter, r *http.Request)
}

type QueueHTTPHandler struct {
	queueService queue.Service
}

func NewQueueHandler() QueueHandler {
	return &QueueHTTPHandler{
		queueService: queue.NewService(),
	}
}

func (h *QueueHTTPHandler) GetPositionHandler(w http.ResponseWriter, r *http.Request) {
	header := request.Header{
		Authorization: r.Header.Get("Authorization"),
	}
	claims, err := jwt.VerifyToken(header.GetToken())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q, err := h.queueService.GetPosition(r.Context(), claims.SessionID)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"position": q.GetPosition(),
	}
	json.NewEncoder(w).Encode(response)

	return
}

func (h *QueueHTTPHandler) EnterHandler(w http.ResponseWriter, r *http.Request) {
	header := request.Header{
		Authorization: r.Header.Get("Authorization"),
	}
	claims, err := jwt.VerifyToken(header.GetToken())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	entered := h.queueService.Enter(r.Context(), claims.SessionID)
	if entered {
		w.WriteHeader(http.StatusOK)

		return
	}

	w.WriteHeader(http.StatusTemporaryRedirect)

	return
}

func (h *QueueHTTPHandler) ExitHandler(w http.ResponseWriter, r *http.Request) {
	header := request.Header{
		Authorization: r.Header.Get("Authorization"),
	}
	claims, err := jwt.VerifyToken(header.GetToken())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	successfullyExited := h.queueService.Exit(r.Context(), claims.SessionID)
	if !successfullyExited {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusOK)

	return
}

func (h *QueueHTTPHandler) GetAverageQueueTimeHandler(w http.ResponseWriter, r *http.Request) {
	averageQueueTime := h.queueService.GetAverageQueueTime(r.Context())

	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"average_queue_time": averageQueueTime,
	}
	json.NewEncoder(w).Encode(response)

	return
}
