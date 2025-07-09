package session

import (
	"context"
	"waitingroom/internal/infra/repositories/session"
)

type Service interface {
	Validate(ctx context.Context, token string) bool
}

type SessionService struct {
	sessionRepository session.Repository
}

func NewService() Service {
	return &SessionService{
		sessionRepository: session.NewRepository(),
	}
}

func (s *SessionService) Validate(ctx context.Context, sessionID string) bool {
	return s.sessionRepository.Exist(ctx, sessionID)
}
