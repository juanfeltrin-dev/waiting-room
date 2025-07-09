package queue

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
	"waitingroom/internal/domain/model"
	"waitingroom/internal/infra/container"
)

const (
	keyQueue = "queue"
)

type Repository interface {
	GetPosition(ctx context.Context, sessionID string) model.Queue
	Enter(ctx context.Context, sessionID string)
	Exit(ctx context.Context, sessionID string)
	First(ctx context.Context) (string, error)
	IsMember(ctx context.Context, sessionID string) bool
}

type CacheRepository struct {
	cache *redis.Client
}

func NewRepository() Repository {
	return &CacheRepository{
		cache: container.GetCache(),
	}
}

func (r *CacheRepository) GetPosition(ctx context.Context, sessionID string) model.Queue {
	position, err := r.cache.ZRank(ctx, "queue", sessionID).Result()
	if err != nil {
		return model.NewQueue(-1)
	}

	return model.NewQueue(position)
}

func (r *CacheRepository) Enter(ctx context.Context, sessionID string) {
	score := float64(time.Now().UnixNano())
	r.cache.ZAdd(ctx, keyQueue, redis.Z{
		Score:  score,
		Member: sessionID,
	})
}

func (r *CacheRepository) Exit(ctx context.Context, sessionID string) {
	r.cache.ZRem(ctx, keyQueue, sessionID)
}

func (r *CacheRepository) First(ctx context.Context) (string, error) {
	ids, err := r.cache.ZRange(ctx, keyQueue, 0, 0).Result()
	if err != nil {
		return "", err
	}

	if len(ids) == 0 {
		return "", errors.New("queue is empty")
	}

	sessionID := ids[0]

	err = r.cache.Watch(ctx, func(tx *redis.Tx) error {
		tx.ZRem(ctx, keyQueue, sessionID)
		return nil
	}, keyQueue)

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (r *CacheRepository) IsMember(ctx context.Context, sessionID string) bool {
	isMember, err := r.cache.SIsMember(ctx, keyQueue, sessionID).Result()
	if err != nil {
		return false
	}

	return isMember
}
