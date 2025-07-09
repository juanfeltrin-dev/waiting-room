package load

import (
	"context"
	"github.com/redis/go-redis/v9"
	"waitingroom/internal/infra/container"
	"waitingroom/internal/model"
)

type Repository interface {
	GetStatus(ctx context.Context) model.Load
	Increment(ctx context.Context, sessionID string)
	Decrement(ctx context.Context, sessionID string)
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

func (r *CacheRepository) GetStatus(ctx context.Context) model.Load {
	count, _ := r.cache.SCard(ctx, "load").Result()

	return model.Load{
		Count: count,
	}
}

func (r *CacheRepository) Increment(ctx context.Context, sessionID string) {
	r.cache.SAdd(ctx, "load", sessionID)
}

func (r *CacheRepository) Decrement(ctx context.Context, sessionID string) {
	r.cache.SRem(ctx, "load", sessionID)
}

func (r *CacheRepository) IsMember(ctx context.Context, sessionID string) bool {
	isMember, err := r.cache.SIsMember(ctx, "load", sessionID).Result()
	if err != nil {
		return false
	}

	return isMember
}
