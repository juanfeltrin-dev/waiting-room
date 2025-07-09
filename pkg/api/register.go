package api

import (
	"github.com/gorilla/mux"
	"waitingroom/internal/infra/protocol/http/middleware"
	v1group "waitingroom/pkg/api/v1"
)

func Register(r *mux.Router) {
	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.Use(middleware.ContentTypeApplicationJsonMiddleware)

	loadGroup := v1group.NewLoadHandler()
	loadPrefix := v1.PathPrefix("/loads").Subrouter()
	loadPrefix.Methods("GET").Path("/status").HandlerFunc(loadGroup.LoadStatusHandler)

	authGroup := v1group.NewAuthHandler()
	authPrefix := v1.PathPrefix("/auth").Subrouter()
	authPrefix.Methods("POST").Path("/login").HandlerFunc(authGroup.LoginHandler)
	authPrefix.Methods("POST").Path("/refresh").HandlerFunc(authGroup.RefreshHandler)

	queueGroup := v1group.NewQueueHandler()
	queuePrefix := v1.PathPrefix("/queues").Subrouter()
	queuePrefix.Use(middleware.ValidateSessionMiddleware)
	queuePrefix.Methods("GET").Path("/position").HandlerFunc(queueGroup.GetPositionHandler)
	queuePrefix.Methods("POST").Path("/enter").HandlerFunc(queueGroup.EnterHandler)
	queuePrefix.Methods("POST").Path("/exit").HandlerFunc(queueGroup.ExitHandler)
	queuePrefix.Methods("GET").Path("/status").HandlerFunc(queueGroup.GetAverageQueueTimeHandler)
}
