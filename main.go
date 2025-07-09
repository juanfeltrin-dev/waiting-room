package main

import (
	"context"
	"waitingroom/cmd/api"
	"waitingroom/internal/infra/container"
)

func main() {
	ctx := context.Background()
	container.Start()
	api.StartServer(ctx)
}
