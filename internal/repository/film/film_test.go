package film

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"testing"
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
	"vk_film/pkg/slices"
)

var testError = errors.New("test error")

type FilmRepositorySuite struct {
	suite.Suite
	filmRepository *PostgresFilm
	mock           sqlxmock.Sqlmock
}

func (frs *FilmRepositorySuite) BeforeEach(t provider.T) {
	db, mock, err := sqlxmock.Newx(sqlxmock.QueryMatcherOption(sqlxmock.QueryMatcherEqual))
	t.Require().NoError(err)
	frs.filmRepository = NewPostgresFilm(db)
	frs.mock = mock
}

func (frs *FilmRepositorySuite) AfterEach(t provider.T) {
	t.Require().NoError(frs.mock.ExpectationsWereMet())
}

func (frs *FilmRepositorySuite) TestGetActorsFunction(t provider.T) {
	t.Title("getActors function")
	t.NewStep("Init test data")

	testFunc := func(db *sqlx.DB, filmId types.Id) ([]Actor, error) {
		tx, err := db.Beginx()
		if err != nil {
			return nil, err
		}

		res, err := getActors(filmId, tx)

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		return res, err
	}

	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	filmId := types.Id(1)

	actorColumns := []string{
		"id", "name", "sex", "birthday",
	}

	actorsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(actorColumns).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time)
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilmActors).WithArgs(filmId).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		flm, err := testFunc(frs.filmRepository.db, filmId)
		t.Require().NoError(err)
		t.Require().EqualValues([]Actor{*actor, *actor, *actor}, flm)
	})

	t.WithNewStep("Postgres error on getFilmActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilmActors).WithArgs(filmId).WillReturnError(testError)
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		_, err := testFunc(frs.filmRepository.db, filmId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows error on getFilmActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilmActors).WithArgs(filmId).WillReturnRows(actorsRows().RowError(1, testError))
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		_, err := testFunc(frs.filmRepository.db, filmId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Incorrect field in row of getFilmActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilmActors).WithArgs(filmId).WillReturnRows(actorsRows().AddRow(1, 1, 1, 1))
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		_, err := testFunc(frs.filmRepository.db, filmId)
		t.Require().Error(err)
	})

	t.WithNewStep("Rows close error on  getFilmActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilmActors).WithArgs(filmId).WillReturnRows(actorsRows().CloseError(testError))
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		_, err := testFunc(frs.filmRepository.db, filmId)
		t.Require().ErrorIs(err, testError)
	})
}

func (frs *FilmRepositorySuite) TestCreateFunction(t provider.T) {
	t.Title("CreateFilm function of Film repository")
	t.NewStep("Init test data")
	film := &Film{
		ID:          1,
		Name:        "Dune",
		Description: "good film",
		DataPublish: time.MustParse("12.03.2003"),
		Rating:      10,
	}

	actorsId := []types.Id{1, 2, 3}

	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	filmColumns := []string{
		"id", "name", "description", "publish_date", "rating",
	}

	actorColumns := []string{
		"id", "name", "sex", "birthday",
	}

	actorsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(actorColumns).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time)
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(createQuery).
			WithArgs(film.Name, film.Description, film.DataPublish.Time, film.Rating).
			WillReturnRows(sqlxmock.NewRows(filmColumns).
				AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating),
			)
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnResult(sqlxmock.NewResult(1, 1))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		flm, err := frs.filmRepository.CreateFilm(film, actorsId)
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: []Actor{*actor, *actor, *actor},
		}, flm)
	})

	t.WithNewStep("Correct actors empty execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(createQuery).
			WithArgs(film.Name, film.Description, film.DataPublish.Time, film.Rating).
			WillReturnRows(sqlxmock.NewRows(filmColumns).
				AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating),
			)
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		flm, err := frs.filmRepository.CreateFilm(film, nil)
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: nil,
		}, flm)
	})

	t.WithNewStep("Conflict user add addActors getFilmActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(createQuery).
			WithArgs(film.Name, film.Description, film.DataPublish.Time, film.Rating).
			WillReturnRows(sqlxmock.NewRows(filmColumns).
				AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating),
			)
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnError(&pq.Error{
			Code:       actorIdConflictCode,
			Constraint: actorIdConstraintName,
		})
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.CreateFilm(film, actorsId)
		t.Require().ErrorIs(err, ErrorActorNotFound)
	})

	t.WithNewStep("Postgres error on begin transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := frs.filmRepository.CreateFilm(film, actorsId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on createFilm query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(createQuery).
			WithArgs(film.Name, film.Description, film.DataPublish.Time, film.Rating).
			WillReturnRows(sqlxmock.NewRows(filmColumns).
				AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating),
			).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.CreateFilm(film, actorsId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on addActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(createQuery).
			WithArgs(film.Name, film.Description, film.DataPublish.Time, film.Rating).
			WillReturnRows(sqlxmock.NewRows(filmColumns).
				AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating),
			)
		frs.mock.ExpectExec(addActors).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.CreateFilm(film, actorsId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on getFilmActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(createQuery).
			WithArgs(film.Name, film.Description, film.DataPublish.Time, film.Rating).
			WillReturnRows(sqlxmock.NewRows(filmColumns).
				AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating),
			)
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnResult(sqlxmock.NewResult(1, 1))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.CreateFilm(film, actorsId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on commit transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(createQuery).
			WithArgs(film.Name, film.Description, film.DataPublish.Time, film.Rating).
			WillReturnRows(sqlxmock.NewRows(filmColumns).
				AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating),
			)
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnResult(sqlxmock.NewResult(1, 1))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := frs.filmRepository.CreateFilm(film, actorsId)
		t.Require().ErrorIs(err, testError)
	})
}

func (frs *FilmRepositorySuite) TestDeleteFunction(t provider.T) {
	t.Title("DeleteFilm function of Film repository")
	t.NewStep("Init test data")
	film := &Film{
		ID:          1,
		Name:        "Dune",
		Description: "good film",
		DataPublish: time.MustParse("12.03.2003"),
		Rating:      10,
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectExec(deleteFilm).
			WithArgs(film.ID).
			WillReturnResult(sqlxmock.NewResult(0, 1))

		t.NewStep("Check result")
		err := frs.filmRepository.DeleteFilm(film.ID)
		t.Require().NoError(err)
	})

	t.WithNewStep("Postgres error for deleteFilm query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectExec(deleteFilm).
			WithArgs(film.ID).WillReturnError(testError)

		t.NewStep("Check result")
		err := frs.filmRepository.DeleteFilm(film.ID)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Row affected error of deleteFilm query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectExec(deleteFilm).
			WithArgs(film.ID).
			WillReturnResult(sqlxmock.NewErrorResult(testError))

		t.NewStep("Check result")
		err := frs.filmRepository.DeleteFilm(film.ID)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Error not found film", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectExec(deleteFilm).
			WithArgs(film.ID).
			WillReturnResult(sqlxmock.NewResult(2, 0))

		t.NewStep("Check result")
		err := frs.filmRepository.DeleteFilm(film.ID)
		t.Require().ErrorIs(err, ErrorFilmNotFound)
	})
}

func (frs *FilmRepositorySuite) TestGetFunction(t provider.T) {
	t.Title("GetFilms function of Film repository")
	t.NewStep("Init test data")

	film := &Film{
		ID:          1,
		Name:        "Dune",
		Description: "good film",
		DataPublish: time.MustParse("12.03.2003"),
		Rating:      10,
	}

	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	filmColumns := []string{
		"id", "name", "description", "publish_date", "rating",
	}

	actorColumns := []string{
		"film_id", "id", "name", "sex", "birthday",
	}

	actorsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(actorColumns).
			AddRow(film.ID, actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(film.ID, actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(film.ID, actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(film.ID+2, actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(film.ID+2, actor.ID, actor.Name, actor.Sex, actor.Birthday.Time)
	}

	filmsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(filmColumns).
			AddRow(film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating).
			AddRow(film.ID+1, film.Name, film.Description, film.DataPublish.Time, film.Rating).
			AddRow(film.ID+2, film.Name, film.Description, film.DataPublish.Time, film.Rating)
	}

	params := Params{
		Order:        types.DESC,
		OrderField:   types.RatingField,
		SearchField:  types.FilmField,
		SearchString: "*",
	}

	t.WithNewStep("Correct rating order desc on film execute", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		films, err := frs.filmRepository.GetFilms(params)
		t.Require().NoError(err)
		t.Require().EqualValues([]FilmWithActors{
			{
				Film: Film{ID: film.ID, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor, *actor},
			},
			{
				Film: Film{ID: film.ID + 1, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{},
			},
			{
				Film: Film{ID: film.ID + 2, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor},
			},
		}, films)
	})

	t.WithNewStep("Correct name order asc on film execute", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, types.NameField, types.ASC)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("a").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		films, err := frs.filmRepository.GetFilms(Params{
			Order:        types.ASC,
			OrderField:   types.NameField,
			SearchField:  types.FilmField,
			SearchString: "a",
		})
		t.Require().NoError(err)
		t.Require().EqualValues([]FilmWithActors{
			{
				Film: Film{ID: film.ID, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor, *actor},
			},
			{
				Film: Film{ID: film.ID + 1, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{},
			},
			{
				Film: Film{ID: film.ID + 2, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor},
			},
		}, films)
	})

	t.WithNewStep("Correct date order asc on film execute", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, types.DataPublishField, types.ASC)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("a").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		films, err := frs.filmRepository.GetFilms(Params{
			Order:        types.ASC,
			OrderField:   types.DataPublishField,
			SearchField:  types.FilmField,
			SearchString: "a",
		})
		t.Require().NoError(err)
		t.Require().EqualValues([]FilmWithActors{
			{
				Film: Film{ID: film.ID, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor, *actor},
			},
			{
				Film: Film{ID: film.ID + 1, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{},
			},
			{
				Film: Film{ID: film.ID + 2, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor},
			},
		}, films)
	})

	t.WithNewStep("Correct rating order desc on actor execute", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchActor, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		films, err := frs.filmRepository.GetFilms(Params{
			Order:        types.DESC,
			OrderField:   types.RatingField,
			SearchField:  types.ActorField,
			SearchString: "*",
		})
		t.Require().NoError(err)
		t.Require().EqualValues([]FilmWithActors{
			{
				Film: Film{ID: film.ID, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor, *actor},
			},
			{
				Film: Film{ID: film.ID + 1, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{},
			},
			{
				Film: Film{ID: film.ID + 2, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor},
			},
		}, films)
	})

	t.WithNewStep("Correct name order asc on actor execute", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchActor, types.NameField, types.ASC)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("a").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		films, err := frs.filmRepository.GetFilms(Params{
			Order:        types.ASC,
			OrderField:   types.NameField,
			SearchField:  types.ActorField,
			SearchString: "a",
		})
		t.Require().NoError(err)
		t.Require().EqualValues([]FilmWithActors{
			{
				Film: Film{ID: film.ID, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor, *actor},
			},
			{
				Film: Film{ID: film.ID + 1, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{},
			},
			{
				Film: Film{ID: film.ID + 2, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor},
			},
		}, films)
	})

	t.WithNewStep("Correct date order asc on actor execute", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchActor, types.DataPublishField, types.ASC)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("a").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		films, err := frs.filmRepository.GetFilms(Params{
			Order:        types.ASC,
			OrderField:   types.DataPublishField,
			SearchField:  types.ActorField,
			SearchString: "a",
		})
		t.Require().NoError(err)
		t.Require().EqualValues([]FilmWithActors{
			{
				Film: Film{ID: film.ID, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor, *actor},
			},
			{
				Film: Film{ID: film.ID + 1, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{},
			},
			{
				Film: Film{ID: film.ID + 2, Name: film.Name,
					Description: film.Description, DataPublish: film.DataPublish, Rating: film.Rating},
				Actors: []Actor{*actor, *actor},
			},
		}, films)
	})

	t.WithNewStep("Correct empty list of films", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(sqlxmock.NewRows(filmColumns))
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		films, err := frs.filmRepository.GetFilms(params)
		t.Require().NoError(err)
		t.Require().EqualValues([]FilmWithActors{}, films)
	})

	t.WithNewStep("Postgres error on commit with empty list of films", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(sqlxmock.NewRows(filmColumns))
		frs.mock.ExpectCommit().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on begin transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows error on getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows().RowError(1, testError))
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Incorrect field in row of getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows().AddRow(1, 1, 1, 1, 1))
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.GetFilms(params)
		t.Require().Error(err)
	})

	t.WithNewStep("Rows close error on getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows().CloseError(testError))
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err = frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows error on getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows().RowError(1, testError))
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err = frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Incorrect field in row of getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows().AddRow(1, 1, 1, 1, 1))
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err = frs.filmRepository.GetFilms(params)
		t.Require().Error(err)
	})

	t.WithNewStep("Rows close error on getFilms query", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows().CloseError(testError))
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err = frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on commit transaction", func(t provider.StepCtx) {
		t.NewStep("Init queries")
		query, args, err := sqlx.In(getFilmsActors, []types.Id{1, 2, 3})
		t.Require().NoError(err)
		query = frs.filmRepository.db.Rebind(query)
		driverArgs := slices.Map(args, func(i interface{}) driver.Value { return i })

		getFilms := fmt.Sprintf(getFilmsSearchFilm, params.OrderField, params.Order)

		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(getFilms).WithArgs("").WillReturnRows(filmsRows())
		frs.mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit().WillReturnError(testError)

		t.NewStep("Check result")
		_, err = frs.filmRepository.GetFilms(params)
		t.Require().ErrorIs(err, testError)
	})
}

func (frs *FilmRepositorySuite) TestUpdateFunction(t provider.T) {
	t.Title("UpdateFilm function of Film repository")
	t.NewStep("Init test data")
	film := &Film{
		ID:          1,
		Name:        "Dune",
		Description: "good film",
		DataPublish: time.MustParse("12.03.2003"),
		Rating:      10,
	}

	actorsId := []types.Id{1, 2, 3}

	actor := &Actor{
		ID:       1,
		Name:     "actor",
		Sex:      types.FEMALE,
		Birthday: time.MustParse("12.03.2003"),
	}

	filmColumns := []string{
		"id", "name", "description", "publish_date", "rating",
	}

	actorColumns := []string{
		"id", "name", "sex", "birthday",
	}

	actorsRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(actorColumns).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time).
			AddRow(actor.ID, actor.Name, actor.Sex, actor.Birthday.Time)
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull(&film.Name),
				getNull(&film.Description),
				sql.NullTime{Valid: true, Time: film.DataPublish.Time},
				sql.NullInt64{Valid: true, Int64: int64(film.Rating)},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectExec(deleteActors).WithArgs(film.ID).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: []Actor{*actor, *actor, *actor},
		}, actors)
	})

	t.WithNewStep("Correct only name execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull(&film.Name),
				getNull((*string)(nil)),
				sql.NullTime{Valid: false},
				sql.NullInt64{Valid: false},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  nil,
			DataPublish:  nil,
			Rating:       nil,
			UpdateActors: false,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: []Actor{*actor, *actor, *actor},
		}, actors)
	})

	t.WithNewStep("Correct only description execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull((*string)(nil)),
				getNull(&film.Description),
				sql.NullTime{Valid: false},
				sql.NullInt64{Valid: false},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         nil,
			Description:  &film.Description,
			DataPublish:  nil,
			Rating:       nil,
			UpdateActors: false,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: []Actor{*actor, *actor, *actor},
		}, actors)
	})

	t.WithNewStep("Correct only publish date execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull((*string)(nil)),
				getNull((*string)(nil)),
				sql.NullTime{Valid: true, Time: film.DataPublish.Time},
				sql.NullInt64{Valid: false},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         nil,
			Description:  nil,
			DataPublish:  &film.DataPublish,
			Rating:       nil,
			UpdateActors: false,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: []Actor{*actor, *actor, *actor},
		}, actors)
	})

	t.WithNewStep("Correct only rating execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull((*string)(nil)),
				getNull((*string)(nil)),
				sql.NullTime{Valid: false},
				sql.NullInt64{Valid: true, Int64: int64(film.Rating)},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         nil,
			Description:  nil,
			DataPublish:  nil,
			Rating:       &film.Rating,
			UpdateActors: false,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: []Actor{*actor, *actor, *actor},
		}, actors)
	})

	t.WithNewStep("Correct only actors Id execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull((*string)(nil)),
				getNull((*string)(nil)),
				sql.NullTime{Valid: false},
				sql.NullInt64{Valid: false},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectExec(deleteActors).WithArgs(film.ID).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit()

		t.NewStep("Check result")
		actors, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         nil,
			Description:  nil,
			DataPublish:  nil,
			Rating:       nil,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().NoError(err)
		t.Require().EqualValues(&FilmWithActors{
			Film:   *film,
			Actors: []Actor{*actor, *actor, *actor},
		}, actors)
	})

	t.WithNewStep("Postgres error on begin transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on updateFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("No film found in updateFilms query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).WillReturnRows(sqlxmock.NewRows(filmColumns))
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, ErrorFilmNotFound)
	})

	t.WithNewStep("Postgres error on deleteActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull(&film.Name),
				getNull(&film.Description),
				sql.NullTime{Valid: true, Time: film.DataPublish.Time},
				sql.NullInt64{Valid: true, Int64: int64(film.Rating)},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectExec(deleteActors).WithArgs(film.ID).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on addActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull(&film.Name),
				getNull(&film.Description),
				sql.NullTime{Valid: true, Time: film.DataPublish.Time},
				sql.NullInt64{Valid: true, Int64: int64(film.Rating)},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectExec(deleteActors).WithArgs(film.ID).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Conflict actor id error on addActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull(&film.Name),
				getNull(&film.Description),
				sql.NullTime{Valid: true, Time: film.DataPublish.Time},
				sql.NullInt64{Valid: true, Int64: int64(film.Rating)},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectExec(deleteActors).WithArgs(film.ID).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).
			WillReturnError(&pq.Error{
				Code:       actorIdConflictCode,
				Constraint: actorIdConstraintName,
			})
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, ErrorActorNotFound)
	})

	t.WithNewStep("Any error on getActors query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull(&film.Name),
				getNull(&film.Description),
				sql.NullTime{Valid: true, Time: film.DataPublish.Time},
				sql.NullInt64{Valid: true, Int64: int64(film.Rating)},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectExec(deleteActors).WithArgs(film.ID).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnError(testError)
		frs.mock.ExpectRollback()

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Postgres error on commit transaction", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		frs.mock.ExpectBegin()
		frs.mock.ExpectQuery(updateFilms).
			WithArgs(film.ID,
				getNull(&film.Name),
				getNull(&film.Description),
				sql.NullTime{Valid: true, Time: film.DataPublish.Time},
				sql.NullInt64{Valid: true, Int64: int64(film.Rating)},
			).
			WillReturnRows(sqlxmock.NewRows(filmColumns).AddRow(
				film.ID, film.Name, film.Description, film.DataPublish.Time, film.Rating,
			))
		frs.mock.ExpectExec(deleteActors).WithArgs(film.ID).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectExec(addActors).WithArgs(film.ID, pq.Array(actorsId)).WillReturnResult(sqlxmock.NewResult(2, 2))
		frs.mock.ExpectQuery(getFilmActors).WithArgs(film.ID).WillReturnRows(actorsRows())
		frs.mock.ExpectCommit().WillReturnError(testError)

		t.NewStep("Check result")
		_, err := frs.filmRepository.UpdateFilm(&UpdateFilm{
			ID:           film.ID,
			Name:         &film.Name,
			Description:  &film.Description,
			DataPublish:  &film.DataPublish,
			Rating:       &film.Rating,
			Actors:       actorsId,
			UpdateActors: true,
		})
		t.Require().ErrorIs(err, testError)
	})
}

func TestRunFilmRepositorySuite(t *testing.T) {
	suite.RunSuite(t, new(FilmRepositorySuite))
}
