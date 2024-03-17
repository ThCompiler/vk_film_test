package film

import (
	"github.com/pkg/errors"
	"vk_film/internal/pkg/types"
)

var (
	ErrorFilmNotFound  = errors.New("film with id not found")
	ErrorActorNotFound = errors.New("actor of film not found")
)

//go:generate mockgen -destination=mocks/repository.go -package=mr -mock_names=Repository=FilmRepository . Repository

type Params struct {
	OrderField   types.OrderField
	Order        types.Order
	SearchField  types.SearchField
	SearchString string
}

type Repository interface {
	// CreateFilm
	// Returns Error:
	//   - SQLError
	//   - ErrorActorNotFound
	CreateFilm(film *Film, actors []types.Id) (*FilmWithActors, error)

	// UpdateFilm
	// Returns Error:
	//   - SQLError
	//   - ErrorFilmNotFound
	//   - ErrorActorNotFound
	UpdateFilm(film *UpdateFilm) (*FilmWithActors, error)

	// DeleteFilm
	// Returns Error:
	//   - SQLError
	//   - ErrorFilmNotFound
	DeleteFilm(id types.Id) error

	// GetFilms
	// Returns Error:
	//   - SQLError
	GetFilms(params Params) ([]FilmWithActors, error)
}
