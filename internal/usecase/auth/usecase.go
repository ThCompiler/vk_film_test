package auth

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"time"
	"vk_film/internal/repository/session"
	"vk_film/internal/repository/user"
)

const (
	ExpiredSessionTime = 48 * time.Hour
)

type SessionManager struct {
	users    user.Repository
	sessions session.Repository
}

func NewSessionManager(users user.Repository, sessions session.Repository) *SessionManager {
	return &SessionManager{
		users:    users,
		sessions: sessions,
	}
}

func (sm *SessionManager) Login(login, password string) (string, error) {
	usr, err := sm.users.GetPasswordByLogin(login)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrorIncorrectPassword
		}
		return "", err
	}

	sessionId := uuid.New().String()

	if err := sm.sessions.Set(sessionId, usr.ID, ExpiredSessionTime); err != nil {
		return "", errors.Wrapf(err, "try save session for user %s", login)
	}

	return sessionId, nil
}

func (sm *SessionManager) Logout(sessionId string) error {
	if err := sm.sessions.Del(sessionId); err != nil {
		return errors.Wrapf(err, "try delete session %s", sessionId)
	}
	return nil
}

func (sm *SessionManager) GetUserId(sessionId string) (*user.User, error) {
	userId, err := sm.sessions.GetUserId(sessionId, ExpiredSessionTime)
	if err != nil {
		return nil, errors.Wrapf(err, "try get session %s", sessionId)
	}

	usr, err := sm.users.GetUserById(userId)
	if err != nil {
		if errors.Is(err, user.ErrorUserNotFound) {
			_ = sm.sessions.Del(sessionId)
			return nil, session.ErrorNoSession
		}
		return nil, errors.Wrapf(err, "try get user by id %d in session %s", userId, sessionId)
	}

	return usr, nil
}
