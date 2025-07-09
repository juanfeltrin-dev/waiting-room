package worker

import (
	"context"
	"log"
	"time"
	"waitingroom/internal/infra/container"
	"waitingroom/internal/services/queue"
)

func Register(ctx context.Context) {
	queueService := queue.NewService()
	go func() {
		for {
			sessionID := queueService.ReleaseEntry(ctx)
			if sessionID != "" {
				log.Println("Chegou sua vez", sessionID)
				container.GetSocket().BroadcastToRoom("/", sessionID, "releaseEntry", "Sua vez chegou!")
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
