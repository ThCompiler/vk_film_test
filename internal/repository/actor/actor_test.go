package actor

import (
	"database/sql"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"testing"
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/film"
)

var testError = errors.New("test error")

type ActorRepositorySuite struct {
	suite.Suite
	actorRepository *PostgresActor
	mock            sqlxmock.Sqlmock
}

func (ars *ActorRepositorySuite) BeforeEach(t provider.T) {
	db, mock, err := sqlxmock.Newx(sqlxmock.QueryMatcherOption(sqlxmock.QueryMatcherEqual))
	t.Require().NoError(err)
	ars.actorRepository = NewPostgresActor(db)
	ars.mock = mock
}

func (ars *ActorRepositorySuite) AfterEach(t provider.T) {
	t.Require().NoError(ars.mock.ExpectationsWereMet())
}

func (ars *ActorRepositorySuite) TestCreateFunction(t provider.T) {
	t.Title("CreateActor function of Actor repository")
	t.NewStep("Init test data")
	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	actorColumns := []string{
		"id", "name", "sex", "birthday",
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectQuery(createQuery).
			WithArgs(actor.Name, actor.Sex, actor.Birthday.Time).
			WillReturnRows(sqlxmock.NewRows(actorColumns).
				AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time),
			)

		t.NewStep("Check result")
		act, err := ars.actorRepository.CreateActor(actor)
		t.Require().NoError(err)
		t.Require().EqualValues(actor, act)
	})

	t.WithNewStep("Postgres error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectQuery(createQuery).
			WithArgs(actor.Name, actor.Sex, actor.Birthday.Time).
			WillReturnError(testError)

		t.NewStep("Check result")
		_, err := ars.actorRepository.CreateActor(actor)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Empty result of execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectQuery(createQuery).
			WithArgs(actor.Name, actor.Sex, actor.Birthday.Time).
			WillReturnRows(sqlxmock.NewRows(actorColumns))

		t.NewStep("Check result")
		_, err := ars.actorRepository.CreateActor(actor)
		t.Require().Error(err)
	})
}

func (ars *ActorRepositorySuite) TestDeleteFunction(t provider.T) {
	t.Title("DeleteActor function of Actor repository")
	t.NewStep("Init test data")
	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectExec(deleteActor).
			WithArgs(actor.ID).
			WillReturnResult(sqlxmock.NewResult(0, 1))

		t.NewStep("Check result")
		err := ars.actorRepository.DeleteActor(actor.ID)
		t.Require().NoError(err)
	})

	t.WithNewStep("Postgres error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectExec(deleteActor).
			WithArgs(actor.ID).
			WillReturnError(testError)

		t.NewStep("Check result")
		err := ars.actorRepository.DeleteActor(actor.ID)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Row affected error of execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectExec(deleteActor).
			WithArgs(actor.ID).
			WillReturnResult(sqlxmock.NewErrorResult(testError))

		t.NewStep("Check result")
		err := ars.actorRepository.DeleteActor(actor.ID)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Error not found actor in execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectExec(deleteActor).
			WithArgs(actor.ID).
			WillReturnResult(sqlxmock.NewResult(2, 0))

		t.NewStep("Check result")
		err := ars.actorRepository.DeleteActor(actor.ID)
		t.Require().ErrorIs(err, ErrorActorNotFound)
	})
}

func (ars *ActorRepositorySuite) TestGetFunction(t provider.T) {
	t.Title("GetActors function of Actor repository")
	t.NewStep("Init test data")
	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	actorColumns := []string{
		"id", "name", "sex", "birthday",
	}

	flm := &film.Film{
		ID:          2,
		Name:        "Dune",
		Description: "good film",
		DataPublish: time.MustParse("12.04.2005"),
		Rating:      10,
	}

	filmColumns := []string{
		"actor_id", "id", "name", "description", "publish_date", "rating",
	}

	actorsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(actorColumns).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID+1, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID+2, actor.Name, actor.Sex, actor.Birthday.Time)
	}

	filmsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(filmColumns).
			AddRow(actor.ID, flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating).
			AddRow(actor.ID, flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating).
			AddRow(actor.ID, flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating).
			AddRow(actor.ID+2, flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating).
			AddRow(actor.ID+2, flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating)
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorsFilms).WillReturnRows(filmsRows())
		ars.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := ars.actorRepository.GetActors()
		t.Require().NoError(err)
		t.Require().EqualValues([]ActorWithFilms{
			{
				Actor: Actor{ID: actor.ID, Name: actor.Name, Sex: actor.Sex, Birthday: actor.Birthday},
				Films: []film.Film{*flm, *flm, *flm},
			},
			{
				Actor: Actor{ID: actor.ID + 1, Name: actor.Name, Sex: actor.Sex, Birthday: actor.Birthday},
				Films: []film.Film{},
			},
			{
				Actor: Actor{ID: actor.ID + 2, Name: actor.Name, Sex: actor.Sex, Birthday: actor.Birthday},
				Films: []film.Film{*flm, *flm},
			},
		}, actors)
	})

	t.WithNewStep("Postgres error on begin transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on getActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnError(testError)
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows error on getActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows().RowError(1, testError))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Incorrect field in row of getActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows().AddRow(1, 1, 1, 1))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().Error(err)
	})

	t.WithNewStep("Rows close error on getActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows().CloseError(testError))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorsFilms).WillReturnError(testError)
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows error on getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorsFilms).WillReturnRows(filmsRows().RowError(1, testError))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows close error on getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorsFilms).WillReturnRows(filmsRows().CloseError(testError))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Incorrect field in row of getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorsFilms).WillReturnRows(filmsRows().AddRow(1, 1, 1, 1, 1, 1))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().Error(err)
	})

	t.WithNewStep("Postgres error on commit transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(getActors).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorsFilms).WillReturnRows(filmsRows())
		ars.mock.ExpectCommit().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := ars.actorRepository.GetActors()
		t.Require().ErrorIs(err, testError)
	})
}

func (ars *ActorRepositorySuite) TestUpdateFunction(t provider.T) {
	t.Title("UpdateActor function of Actor repository")
	t.NewStep("Init test data")
	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	actorColumns := []string{
		"id", "name", "sex", "birthday",
	}

	flm := &film.Film{
		ID:          2,
		Name:        "Dune",
		Description: "good film",
		DataPublish: time.MustParse("12.04.2005"),
		Rating:      10,
	}

	filmColumns := []string{
		"id", "name", "description", "publish_date", "rating",
	}

	actorsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(actorColumns).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time)
	}

	filmsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(filmColumns).
			AddRow(flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating).
			AddRow(flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating).
			AddRow(flm.ID, flm.Name, flm.Description, flm.DataPublish.Time, flm.Rating)
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(&actor.Name),
				getNullString((*string)(&actor.Sex)),
				sql.NullTime{Valid: true, Time: actor.Birthday.Time},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows())
		ars.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     &actor.Name,
			Sex:      &actor.Sex,
			Birthday: &actor.Birthday,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&ActorWithFilms{
			Actor: Actor{ID: actor.ID, Name: actor.Name, Sex: actor.Sex, Birthday: actor.Birthday},
			Films: []film.Film{*flm, *flm, *flm},
		}, actors)
	})

	t.WithNewStep("Correct only name execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(&actor.Name),
				getNullString(nil),
				sql.NullTime{Valid: false},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows())
		ars.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     &actor.Name,
			Sex:      nil,
			Birthday: nil,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&ActorWithFilms{
			Actor: Actor{ID: actor.ID, Name: actor.Name, Sex: actor.Sex, Birthday: actor.Birthday},
			Films: []film.Film{*flm, *flm, *flm},
		}, actors)
	})

	t.WithNewStep("Correct only sex execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(nil),
				getNullString((*string)(&actor.Sex)),
				sql.NullTime{Valid: false},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows())
		ars.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      &actor.Sex,
			Birthday: nil,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&ActorWithFilms{
			Actor: Actor{ID: actor.ID, Name: actor.Name, Sex: actor.Sex, Birthday: actor.Birthday},
			Films: []film.Film{*flm, *flm, *flm},
		}, actors)
	})

	t.WithNewStep("Correct only time execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(nil),
				getNullString(nil),
				sql.NullTime{Valid: true, Time: actor.Birthday.Time},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows())
		ars.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&ActorWithFilms{
			Actor: Actor{ID: actor.ID, Name: actor.Name, Sex: actor.Sex, Birthday: actor.Birthday},
			Films: []film.Film{*flm, *flm, *flm},
		}, actors)
	})

	t.WithNewStep("Postgres error on begin transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on updateActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).WillReturnError(testError)
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("No actor found in updateActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).WillReturnRows(sqlxmock.NewRows(actorColumns))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().ErrorIs(err, ErrorActorNotFound)
	})

	t.WithNewStep("Rows error on getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(nil),
				getNullString(nil),
				sql.NullTime{Valid: true, Time: actor.Birthday.Time},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows().RowError(1, testError))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(nil),
				getNullString(nil),
				sql.NullTime{Valid: true, Time: actor.Birthday.Time},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnError(testError)
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows close on getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(nil),
				getNullString(nil),
				sql.NullTime{Valid: true, Time: actor.Birthday.Time},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows().CloseError(testError))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Incorrect field in row of getActorFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(nil),
				getNullString(nil),
				sql.NullTime{Valid: true, Time: actor.Birthday.Time},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows().AddRow(1, 1, 1, 1, 1))
		ars.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().Error(err)
	})

	t.WithNewStep("Postgres error on commit transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ars.mock.ExpectBegin()
		ars.mock.ExpectQuery(updateActors).
			WithArgs(actor.ID,
				getNullString(nil),
				getNullString(nil),
				sql.NullTime{Valid: true, Time: actor.Birthday.Time},
			).WillReturnRows(actorsRows())
		ars.mock.ExpectQuery(getActorFilms).WillReturnRows(filmsRows())
		ars.mock.ExpectCommit().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := ars.actorRepository.UpdateActor(&UpdateActor{
			ID:       actor.ID,
			Name:     nil,
			Sex:      nil,
			Birthday: &actor.Birthday,
		})
		t.Require().ErrorIs(err, testError)
	})
}

func TestRunActorRepositorySuite(t *testing.T) {
	suite.RunSuite(t, new(ActorRepositorySuite))
}
