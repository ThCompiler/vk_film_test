package film

import (
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
)

type UpdateFilm struct {
	ID           types.Id
	Name         *string
	Description  *string
	DataPublish  *time.FormattedTime
	Rating       *types.Rating
	Actors       []types.Id
	UpdateActors bool
}

type Film struct {
	ID          types.Id
	Name        string
	Description string
	DataPublish time.FormattedTime
	Rating      types.Rating
}

type FilmWithActors struct {
	Film
	Actors []Actor
}

type Actor struct {
	ID       types.Id           `json:"id"`
	Name     string             `json:"name"`
	Sex      types.Sexes        `json:"sex"`
	Birthday time.FormattedTime `json:"birthday"`
}
