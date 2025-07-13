package queue

import (
	"context"
	"errors"
	"waitingroom/internal/infra/repositories/load"
	"waitingroom/internal/infra/repositories/queue"
	"waitingroom/internal/infra/repositories/session"
	"waitingroom/internal/model"
)

type Service interface {
	Enter(ctx context.Context, sessionID string) bool
	GetPosition(ctx context.Context, sessionID string) (model.Queue, error)
	Exit(ctx context.Context, sessionID string) bool
	ReleaseEntry(ctx context.Context) string
	GetAverageQueueTime(ctx context.Context) int64
}

type QueueService struct {
	LoadRepository    load.Repository
	QueueRepository   queue.Repository
	SessionRepository session.Repository
}

func NewService() Service {
	return &QueueService{
		LoadRepository:    load.NewRepository(),
		QueueRepository:   queue.NewRepository(),
		SessionRepository: session.NewRepository(),
	}
}

func (s *QueueService) GetPosition(ctx context.Context, sessionID string) (model.Queue, error) {
	queueModel := s.QueueRepository.GetPosition(ctx, sessionID)
	if queueModel.OutOfQueue() {
		return model.Queue{}, errors.New("user is not in queue")
	}

	return queueModel, nil
}

func (s *QueueService) Enter(ctx context.Context, sessionID string) bool {
	isMemberInQueue := s.QueueRepository.IsMember(ctx, sessionID)
	if isMemberInQueue {
		return true
	}

	isMemberInLoad := s.LoadRepository.IsMember(ctx, sessionID)
	if isMemberInLoad {
		return false
	}

	loadModel := s.LoadRepository.GetStatus(ctx)
	if loadModel.ReachedCapacity() {
		err := s.SessionRepository.Init(ctx, sessionID)
		if err != nil {
			return false
		}

		s.QueueRepository.Enter(ctx, sessionID)

		return true
	}

	s.LoadRepository.Increment(ctx, sessionID)

	return false
}

func (s *QueueService) Exit(ctx context.Context, sessionID string) bool {
	isMember := s.LoadRepository.IsMember(ctx, sessionID)
	if !isMember {
		return false
	}

	s.LoadRepository.Decrement(ctx, sessionID)

	return true
}

func (s *QueueService) ReleaseEntry(ctx context.Context) string {
	loadModel := s.LoadRepository.GetStatus(ctx)
	if loadModel.ReachedCapacity() {
		return ""
	}

	sessionID, _ := s.QueueRepository.First(ctx)
	if sessionID != "" {
		s.LoadRepository.Increment(ctx, sessionID)
		s.QueueRepository.Exit(ctx, sessionID)
		s.SessionRepository.Exit(ctx, sessionID)

		return sessionID
	}

	return ""
}

func (s *QueueService) GetAverageQueueTime(ctx context.Context) int64 {
	return s.SessionRepository.GetAverageQueueTime(ctx)
}
