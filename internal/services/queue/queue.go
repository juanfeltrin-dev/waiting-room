package queue

import (
	"context"
	"errors"
	"waitingroom/internal/domain/model"
	"waitingroom/internal/infra/repositories/load"
	"waitingroom/internal/infra/repositories/queue"
	"waitingroom/internal/infra/repositories/session"
)

type Service interface {
	Enter(ctx context.Context, sessionID string) bool
	GetPosition(ctx context.Context, sessionID string) (model.Queue, error)
	Exit(ctx context.Context, sessionID string) bool
	ReleaseEntry(ctx context.Context) string
	GetAverageQueueTime(ctx context.Context) int64
}

type QueueService struct {
	loadRepository    load.Repository
	queueRepository   queue.Repository
	sessionRepository session.Repository
}

func NewService() Service {
	return &QueueService{
		loadRepository:    load.NewRepository(),
		queueRepository:   queue.NewRepository(),
		sessionRepository: session.NewRepository(),
	}
}

func (s *QueueService) GetPosition(ctx context.Context, sessionID string) (model.Queue, error) {
	queueModel := s.queueRepository.GetPosition(ctx, sessionID)
	if queueModel.OutOfQueue() {
		return model.Queue{}, errors.New("user is not in queue")
	}

	return queueModel, nil
}

func (s *QueueService) Enter(ctx context.Context, sessionID string) bool {
	isMemberInQueue := s.queueRepository.IsMember(ctx, sessionID)
	if isMemberInQueue {
		return true
	}

	isMemberInLoad := s.loadRepository.IsMember(ctx, sessionID)
	if isMemberInLoad {
		return false
	}

	loadModel := s.loadRepository.GetStatus(ctx)
	if loadModel.ReachedCapacity() {
		err := s.sessionRepository.Init(ctx, sessionID)
		if err != nil {
			return false
		}

		s.queueRepository.Enter(ctx, sessionID)

		return true
	}

	s.loadRepository.Increment(ctx, sessionID)

	return false
}

func (s *QueueService) Exit(ctx context.Context, sessionID string) bool {
	isMember := s.loadRepository.IsMember(ctx, sessionID)
	if !isMember {
		return false
	}

	s.loadRepository.Decrement(ctx, sessionID)

	return true
}

func (s *QueueService) ReleaseEntry(ctx context.Context) string {
	loadModel := s.loadRepository.GetStatus(ctx)
	if loadModel.ReachedCapacity() {
		return ""
	}

	sessionID, _ := s.queueRepository.First(ctx)
	if sessionID != "" {
		s.loadRepository.Increment(ctx, sessionID)
		s.queueRepository.Exit(ctx, sessionID)
		s.sessionRepository.Exit(ctx, sessionID)

		return sessionID
	}

	return ""
}

func (s *QueueService) GetAverageQueueTime(ctx context.Context) int64 {
	return s.sessionRepository.GetAverageQueueTime(ctx)
}
