package response

import (
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/actor"
	"vk_film/internal/repository/film"
	"vk_film/pkg/slices"
)

type Actor struct {
	ID       types.Id           `json:"id" swaggertype:"integer" format:"uint64" example:"5"`
	Name     string             `json:"name" swaggertype:"string" example:"Тимоти Шаламе"`
	Sex      string             `json:"sex" swaggertype:"string" example:"male" enums:"male,female"`
	Birthday time.FormattedTime `json:"birthday" swaggertype:"string" format:"date" example:"12.02.2002"`
}

type ActorWithFilms struct {
	Actor
	Films []ActorFilms `json:"films,omitempty"`
}

type ActorFilms struct {
	ID          types.Id           `json:"id" swaggertype:"integer" format:"uint64" example:"5"`
	Name        string             `json:"name" swaggertype:"string" example:"Dune"`
	Description string             `json:"description" swaggertype:"string" example:"Futuristic film"`
	DataPublish time.FormattedTime `json:"data_publish" swaggertype:"string" format:"date" example:"12.02.2023"`
	Rating      types.Rating       `json:"rating" swaggertype:"integer" format:"uint8" example:"9"`
}

func FromRepositoryActorsWithFilms(actorsRepository []actor.ActorWithFilms) []ActorWithFilms {
	return slices.Map(actorsRepository, func(act actor.ActorWithFilms) ActorWithFilms {
		return *FromRepositoryActorWithFilms(&act)
	})
}

func FromRepositoryActorWithFilms(actorRepository *actor.ActorWithFilms) *ActorWithFilms {
	return &ActorWithFilms{
		Actor: Actor{
			ID:       actorRepository.ID,
			Name:     actorRepository.Name,
			Sex:      string(actorRepository.Sex),
			Birthday: actorRepository.Birthday,
		},
		Films: slices.Map(actorRepository.Films, func(flm film.Film) ActorFilms {
			return ActorFilms{
				ID:          flm.ID,
				Name:        flm.Name,
				Description: flm.Description,
				DataPublish: flm.DataPublish,
				Rating:      flm.Rating,
			}
		}),
	}
}
