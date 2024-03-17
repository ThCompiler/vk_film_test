package session

import (
	"github.com/pkg/errors"
	"time"
	"vk_film/internal/pkg/types"
)

var ErrorNoSession = errors.New("session was not found")

//go:generate mockgen -destination=mocks/repository.go -package=mr -mock_names=Repository=SessionRepository . Repository

type Repository interface {
	Set(sessionId string, userId types.Id, expiredTime time.Duration) error
	GetUserId(sessionId string, updateExpiredTime time.Duration) (types.Id, error)
	Del(sessionId string) error
}
