package session

import (
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"time"
	"vk_film/internal/pkg/types"
)

type RedisSession struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisSession(client *redis.Client) *RedisSession {
	return &RedisSession{client: client, ctx: context.Background()}
}

func (rs *RedisSession) Set(sessionId string, userId types.Id, expiredTime time.Duration) error {
	if err := rs.client.Set(rs.ctx, sessionId, uint64(userId), expiredTime).Err(); err != nil {
		return errors.Wrapf(err,
			"error when try create session with uniqId: %s, and userId: %d", sessionId, userId)
	}
	return nil
}

func (rs *RedisSession) GetUserId(sessionId string, updateExpiredTime time.Duration) (types.Id, error) {
	userId, err := rs.client.Get(rs.ctx, sessionId).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = ErrorNoSession
		}

		return 0, errors.Wrapf(err,
			"error when try found session with sessionId: %s", sessionId)
	}

	if err = rs.client.Expire(rs.ctx, sessionId, updateExpiredTime).Err(); err != nil {
		return 0, errors.Wrapf(err,
			"error when try update expired time with sessionId: %s", sessionId)
	}
	return types.Id(userId), nil
}

func (rs *RedisSession) Del(sessionId string) error {
	if err := rs.client.Del(rs.ctx, sessionId).Err(); err != nil {
		return errors.Wrapf(err,
			"error when try delete session with sessionId: %s", sessionId)
	}
	return nil
}
