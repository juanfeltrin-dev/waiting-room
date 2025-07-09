package session

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"waitingroom/internal/infra/container"
)

const (
	key                 = "session:"
	keyAverageQueueTime = "avg_queue_time"
	ttl                 = time.Hour * 2
)

type Repository interface {
	Exist(ctx context.Context, sessionID string) bool
	Init(ctx context.Context, sessionID string) error
	Exit(ctx context.Context, sessionID string)
	GetAverageQueueTime(ctx context.Context) int64
}

type CacheRepository struct {
	cache *redis.Client
}

func NewRepository() Repository {
	return &CacheRepository{
		cache: container.GetCache(),
	}
}

func (r *CacheRepository) Exist(ctx context.Context, sessionID string) bool {
	_, err := r.cache.Get(ctx, key+sessionID).Result()
	if err != nil {
		return false
	}

	return true
}

func (r *CacheRepository) Init(ctx context.Context, sessionID string) error {
	_, err := r.cache.SetEx(ctx, key+sessionID, time.Now().Unix(), ttl).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *CacheRepository) Exit(ctx context.Context, sessionID string) {
	sessionEntranceTime, err := r.cache.Get(ctx, key+sessionID).Result()
	if err != nil {
		return
	}

	sessionEntranceTimeConvert, err := strconv.ParseInt(sessionEntranceTime, 10, 64)
	if err != nil {
		return
	}

	duration := time.Now().Unix() - time.Unix(sessionEntranceTimeConvert, 0).Unix()

	r.cache.XAdd(ctx, &redis.XAddArgs{
		Stream: keyAverageQueueTime,
		Values: map[string]interface{}{
			"token":    sessionID,
			"duration": duration,
		},
	})
	r.cache.Del(ctx, key+sessionID)
}

func (r *CacheRepository) GetAverageQueueTime(ctx context.Context) int64 {
	durationsInQueue, err := r.cache.XRevRangeN(ctx, keyAverageQueueTime, "+", "-", 100).Result()
	if err != nil {
		return 0
	}

	var soma int64
	for _, durationInQueue := range durationsInQueue {
		soma += durationInQueue.Values["duration"].(int64)
	}

	return soma / int64(len(durationsInQueue))
}
