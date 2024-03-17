package response

import (
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/film"
	"vk_film/pkg/slices"
)

type Film struct {
	ID          types.Id           `json:"id" swaggertype:"integer" format:"uint64" example:"5"`
	Name        string             `json:"name" swaggertype:"string" example:"Dune"`
	Description string             `json:"description" swaggertype:"string" example:"Futuristic film"`
	DataPublish time.FormattedTime `json:"data_publish" swaggertype:"string" format:"date" example:"12.02.2023"`
	Rating      types.Rating       `json:"rating" swaggertype:"integer" format:"uint8" example:"9"`
	Actors      []FilmActors       `json:"actors,omitempty"`
}

type FilmActors struct {
	Actor
}

func FromRepositoryFilmsWithActor(filmsRepository []film.FilmWithActors) []Film {
	return slices.Map(filmsRepository, func(flm film.FilmWithActors) Film {
		return *FromRepositoryFilmWithActor(&flm)
	})
}

func FromRepositoryFilmWithActor(filmRepository *film.FilmWithActors) *Film {
	return &Film{
		ID:          filmRepository.ID,
		Name:        filmRepository.Name,
		Description: filmRepository.Description,
		DataPublish: filmRepository.DataPublish,
		Rating:      filmRepository.Rating,
		Actors: slices.Map(filmRepository.Actors, func(act film.Actor) FilmActors {
			return FilmActors{
				Actor: Actor{
					ID:       act.ID,
					Name:     act.Name,
					Sex:      string(act.Sex),
					Birthday: act.Birthday,
				},
			}
		}),
	}
}
