package handlers

import "github.com/pkg/errors"

var (
	ErrorIncorrectLoginOrPassword = errors.New("incorrect login or password")
	ErrorCannotReadBody           = errors.New("can't read body")
	ErrorIncorrectBodyContent     = errors.New("incorrect body content")
	ErrorUserNotPermitted         = errors.New("the user with the current role does not have enough permissions")
	ErrorUnknownError             = errors.New("unknown error, try again later")
	ErrorIncorrectQueryParam      = errors.New("invalid query parameter")

	ErrorUserAlreadyExists = errors.New("user already exists")
	ErrorActorNotFound     = errors.New("actor not found")
	ErrorFilmNotFound      = errors.New("film not found")
	ErrorUserNotFound      = errors.New("user not found")
)
