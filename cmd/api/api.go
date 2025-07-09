package api

import (
	"context"
	"flag"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"waitingroom/internal/infra/container"
	httpprotocol "waitingroom/internal/infra/protocol/http"
	"waitingroom/pkg/api"
	"waitingroom/pkg/socket"
	"waitingroom/pkg/worker"
)

func StartServer(ctx context.Context) {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	handler := httpprotocol.StartServer()
	api.Register(handler)
	socket.Register(handler)
	worker.Register(ctx)
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:63342"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	srv := &http.Server{
		Handler: crs.Handler(handler),
		Addr:    ":8000",
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Println("server started by")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-c

	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("Server Shutdown error:", err)
	}

	err = container.GetSocket().Close()
	if err != nil {
		log.Fatal("socket shutdown failed:", err)
	}

	log.Println("shutting down")
	os.Exit(0)
}
