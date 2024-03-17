package middleware

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
	"time"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/session"
	"vk_film/internal/repository/user"
	"vk_film/internal/usecase/auth"
	"vk_film/pkg/mux"
)

const (
	UserField    = types.ContextField("user_info")
	SessionField = types.ContextField("session_id")
)

func updateCookie(w http.ResponseWriter, cook *http.Cookie) {
	cook.Expires = time.Now().Add(auth.ExpiredSessionTime)
	cook.Path = "/"
	cook.HttpOnly = true
	http.SetCookie(w, cook)
}

func clearCookie(w http.ResponseWriter, cook *http.Cookie) {
	cook.Expires = time.Now().AddDate(0, 0, -1)
	cook.Path = "/"
	cook.HttpOnly = true
	http.SetCookie(w, cook)
}

func CheckSession(sessionManager auth.Manager) mux.MiddlewareFunc {
	return func(fun mux.ExtendedHandleFunc) mux.ExtendedHandleFunc {
		return func(w http.ResponseWriter, r *http.Request, params mux.Params) {
			sessionCookie, err := r.Cookie(string(SessionField))
			if err != nil {
				GetLogger(r).Warn("in parsing cookie: %s", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sessionID := sessionCookie.Value
			res, err := sessionManager.GetUserId(sessionID)
			if err != nil {
				if errors.Is(err, session.ErrorNoSession) {
					GetLogger(r).Debug("no session by id %s", sessionID)
					clearCookie(w, sessionCookie)
					w.WriteHeader(http.StatusUnauthorized)
				} else {
					GetLogger(r).Error(errors.Wrapf(err, "error with session id %s", sessionID))
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}

			GetLogger(r).Debug("get session for user: %d", res.ID)
			updateCookie(w, sessionCookie)

			contextWithFields := context.WithValue(r.Context(), UserField, res)
			contextedRequest := r.WithContext(context.WithValue(contextWithFields, SessionField, sessionID))
			// Process request
			fun(w, contextedRequest, params)
		}
	}
}

func CheckNoSession(sessionManager auth.Manager) mux.MiddlewareFunc {
	return func(fun mux.ExtendedHandleFunc) mux.ExtendedHandleFunc {
		return func(w http.ResponseWriter, r *http.Request, params mux.Params) {
			sessionCookie, err := r.Cookie(string(SessionField))
			if err != nil {
				fun(w, r, params)
				return
			}

			sessionID := sessionCookie.Value
			res, err := sessionManager.GetUserId(sessionID)
			if err != nil {
				clearCookie(w, sessionCookie)
				fun(w, r, params)
				return
			}

			GetLogger(r).Debug("user already authorized: %d", res.ID)
			updateCookie(w, sessionCookie)

			w.WriteHeader(http.StatusTeapot)
		}
	}
}

func GetUser(r *http.Request) *user.User {
	if lg := r.Context().Value(UserField); lg != nil {
		return lg.(*user.User)
	}

	return nil
}

func GetSession(r *http.Request) *string {
	if lg := r.Context().Value(SessionField); lg != nil {
		if sessionId, ok := lg.(string); ok {
			return &sessionId
		}
		return nil
	}

	return nil
}
