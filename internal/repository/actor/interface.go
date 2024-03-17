package actor

import (
	"github.com/pkg/errors"
	"vk_film/internal/pkg/types"
)

var (
	ErrorActorNotFound = errors.New("actor with id not found")
)

//go:generate mockgen -destination=mocks/repository.go -package=mr -mock_names=Repository=ActorRepository . Repository

type Repository interface {
	// CreateActor
	// Returns Error:
	//   - SQLError
	CreateActor(actor *Actor) (*Actor, error)

	// UpdateActor
	// Returns Error:
	//   - SQLError
	//   - ErrorActorNotFound
	UpdateActor(actor *UpdateActor) (*ActorWithFilms, error)

	// DeleteActor
	// Returns Error:
	//   - SQLError
	//   - ErrorActorNotFound
	DeleteActor(id types.Id) error

	// GetActors
	// Returns Error:
	//   - SQLError
	GetActors() ([]ActorWithFilms, error)
}
