package user

import (
	"github.com/pkg/errors"
	"vk_film/internal/pkg/types"
)

var (
	ErrorUserNotFound       = errors.New("user not found")
	ErrorLoginAlreadyExists = errors.New("user with this login already exists")
)

//go:generate mockgen -destination=mocks/repository.go -package=mr -mock_names=Repository=UserRepository . Repository

type Repository interface {
	// CreateUser
	// Returns Error:
	//   - SQLError
	//   - ErrorLoginAlreadyExists
	CreateUser(user *User) (*User, error)

	// UpdateUserRole
	// Returns Error:
	//   - SQLError
	//   - ErrorUserNotFound
	UpdateUserRole(user *User) (*User, error)

	// DeleteUser
	// Returns Error:
	//   - SQLError
	//   - ErrorUserNotFound
	DeleteUser(id types.Id) error

	// GetPasswordByLogin
	// Returns Error:
	//   - SQLError
	//   - ErrorUserNotFound
	GetPasswordByLogin(login string) (*LoginUser, error)

	// GetUserById
	// Returns Error:
	//   - SQLError
	//   - ErrorUserNotFound
	GetUserById(id types.Id) (*User, error)

	// GetUsers
	// Returns Error:
	//   - SQLError
	GetUsers() ([]User, error)
}
