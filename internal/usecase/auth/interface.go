package auth

import (
	"github.com/pkg/errors"
	"vk_film/internal/repository/user"
)

var (
	ErrorIncorrectPassword = errors.New("incorrect password")
)

//go:generate mockgen -destination=mocks/manager.go -package=mu -mock_names=Manager=SessionManager . Manager

type Manager interface {
	Login(login, password string) (string, error)
	Logout(sessionId string) error
	GetUserId(sessionId string) (*user.User, error)
}
