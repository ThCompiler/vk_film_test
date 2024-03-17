package actor

import (
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/film"
)

type UpdateActor struct {
	ID       types.Id
	Name     *string
	Sex      *types.Sexes
	Birthday *time.FormattedTime
}

type Actor struct {
	ID       types.Id
	Name     string
	Sex      types.Sexes
	Birthday time.FormattedTime
}

type ActorWithFilms struct {
	Actor
	Films []film.Film
}
