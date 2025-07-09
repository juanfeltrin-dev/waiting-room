package load

import (
	"context"
	"waitingroom/internal/infra/repositories/load"
	"waitingroom/internal/model"
)

type Service interface {
	GetStatus(ctx context.Context) model.Load
}

type LoadService struct {
	Repository load.Repository
}

func NewService() Service {
	return &LoadService{
		Repository: load.NewRepository(),
	}
}

func (s *LoadService) GetStatus(ctx context.Context) model.Load {
	return s.Repository.GetStatus(ctx)
}
