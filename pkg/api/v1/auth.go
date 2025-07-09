package v1

import (
	"encoding/json"
	"log"
	"net/http"
	"waitingroom/internal/infra/jwt"
	"waitingroom/internal/schemas/auth"
	"waitingroom/internal/schemas/request"
	"waitingroom/internal/services/session"
)

type AuthHandler interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	RefreshHandler(w http.ResponseWriter, r *http.Request)
}

type AuthHTTPHandler struct {
	sessionService session.Service
}

func NewAuthHandler() AuthHandler {
	return &AuthHTTPHandler{
		sessionService: session.NewService(),
	}
}

func (h *AuthHTTPHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := auth.NewSessionID()
	token, err := jwt.CreateToken(sessionID)
	if err != nil {
		log.Println("error on create token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"token": token,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHTTPHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	header := request.Header{
		Authorization: r.Header.Get("Authorization"),
	}

	claims, err := jwt.ParseUnverifiedToken(header.GetToken())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists := h.sessionService.Validate(r.Context(), claims.SessionID)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := jwt.CreateToken(claims.SessionID)
	if err != nil {
		log.Println("error on create token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"token": token,
	}
	json.NewEncoder(w).Encode(response)
}
