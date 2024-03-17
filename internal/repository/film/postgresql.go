package film

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"vk_film/internal/pkg/types"
)

const (
	createQuery = `
		INSERT INTO films (name, description, publish_date, rating)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, publish_date, rating
	`

	addActors = `
		INSERT INTO film_actor (film_id, actor_id)
		SELECT $1, actor 
		FROM unnest($2::int[]) as actor		
	`

	deleteFilm = `
		DELETE FROM films WHERE id = $1
	`

	updateFilms = `
		UPDATE films SET name = upd_film.upd_name, description = upd_film.upd_description, 
		                 publish_date = upd_film.upd_publish_date, rating = upd_film.upd_rating
			FROM (
				SELECT COALESCE($2, films.name) as upd_name, 
					   COALESCE($3, films.description) as upd_description, 
					   COALESCE($4, films.publish_date) as upd_publish_date,
					   COALESCE($5, films.rating) as upd_rating 
				FROM films WHERE id = $1
			) as upd_film
			WHERE id = $1
			RETURNING id, name, description, publish_date, rating
	`

	deleteActors = `
		DELETE FROM film_actor WHERE film_id = $1
	`

	getFilmActors = `
		SELECT actors.id, actors.name, actors.sex, actors.birthday FROM film_actor
			JOIN actors on (actors.id = film_actor.actor_id)
		 	WHERE film_actor.film_id = $1
	`

	getFilmsSearchFilm = `
		SELECT id, name, description, publish_date, rating FROM films 
		WHERE name LIKE '%%' || $1 || '%%'
		ORDER BY %s %s
	`

	getFilmsSearchActor = `
		SELECT films.id, films.name, films.description, films.publish_date, films.rating FROM films 
		JOIN film_actor on (films.id = film_actor.film_id)
			JOIN actors on (actors.id = film_actor.actor_id)
		WHERE actors.name LIKE '%%' || $1 || '%%'
		ORDER BY films.%s %s
	`

	getFilmsActors = `
		SELECT films.id, actors.id, actors.name, actors.sex, actors.birthday FROM films 
			JOIN film_actor on (films.id = film_actor.film_id)
			JOIN actors on (actors.id = film_actor.actor_id)
			WHERE films.id in (?)
	`
)

type PostgresFilm struct {
	db *sqlx.DB
}

func NewPostgresFilm(db *sqlx.DB) *PostgresFilm {
	return &PostgresFilm{
		db: db,
	}
}

var _ = Repository(&PostgresFilm{})

func getActors(filmId types.Id, tx *sqlx.Tx) ([]Actor, error) {
	rows, err := tx.Queryx(getFilmActors, filmId)
	if err != nil {
		return nil, errors.Wrapf(err, "can't execute get query actors for film with id %d", filmId)
	}

	actors := make([]Actor, 0)

	for rows.Next() {
		var filmActor Actor

		err := rows.Scan(
			&filmActor.ID,
			&filmActor.Name,
			&filmActor.Sex,
			&filmActor.Birthday,
		)

		if err != nil {
			return nil, errors.Wrapf(err, "can't scan get actors for film with id %d", filmId)
		}

		actors = append(actors, filmActor)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "can't scan get actors for film with id %d", filmId)
	}

	return actors, nil
}

func (pf *PostgresFilm) CreateFilm(film *Film, actors []types.Id) (*FilmWithActors, error) {
	tx, err := pf.db.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "can't create transaction for create film")
	}

	newFilm := &FilmWithActors{}
	if err := tx.QueryRowx(createQuery, film.Name, film.Description, &film.DataPublish, film.Rating).
		Scan(
			&newFilm.ID,
			&newFilm.Name,
			&newFilm.Description,
			&newFilm.DataPublish,
			&newFilm.Rating,
		); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't create film")
	}

	if len(actors) != 0 {
		if _, err := tx.Exec(addActors, newFilm.ID, pq.Array(actors)); err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(checkActorConflictError(err), "can't create actors for film")
		}

		newFilm.Actors, err = getActors(newFilm.ID, tx)
		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "can't get actors for film")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "can't commit transaction for create film")
	}

	return newFilm, nil
}

func getNull[T any](value *T) sql.Null[T] {
	if value == nil {
		return sql.Null[T]{Valid: false}
	}
	return sql.Null[T]{Valid: true, V: *value}
}

func (pf *PostgresFilm) UpdateFilm(film *UpdateFilm) (*FilmWithActors, error) {
	name := getNull(film.Name)
	description := getNull(film.Description)

	rating := sql.NullInt64{Valid: false}
	if film.Rating != nil {
		rating = sql.NullInt64{Valid: true, Int64: int64(*film.Rating)}
	}

	dataPublish := sql.NullTime{Valid: false}
	if film.DataPublish != nil {
		dataPublish = sql.NullTime{Valid: true, Time: film.DataPublish.Time}
	}

	tx, err := pf.db.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "can't create transaction for update film")
	}

	updatedFilm := &FilmWithActors{}
	if err := tx.QueryRowx(updateFilms, film.ID, name, description, dataPublish, rating).
		Scan(
			&updatedFilm.ID,
			&updatedFilm.Name,
			&updatedFilm.Description,
			&updatedFilm.DataPublish,
			&updatedFilm.Rating,
		); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorFilmNotFound
		}
		return nil, errors.Wrapf(err, "can't update film with id %d", film.ID)
	}

	// Обновление данных об актёрах
	if film.UpdateActors {
		if _, err := tx.Exec(deleteActors, updatedFilm.ID); err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrapf(err, "can't delete old actors for updated film with id %d", film.ID)
		}

		if len(film.Actors) != 0 {
			if _, err := tx.Exec(addActors, updatedFilm.ID, pq.Array(film.Actors)); err != nil {
				_ = tx.Rollback()
				return nil, errors.Wrapf(checkActorConflictError(err),
					"can't create actors for updated film with id %d", film.ID)
			}
		}
	}

	// Получаем список фильмов для автора
	updatedFilm.Actors, err = getActors(updatedFilm.ID, tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrapf(err, "can't get actors for updated film with id %d", film.ID)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrapf(err, "can't commit transaction for update film with id %d", film.ID)
	}

	return updatedFilm, nil
}

func (pf *PostgresFilm) DeleteFilm(id types.Id) error {
	res, err := pf.db.Exec(deleteFilm, id)
	if err != nil {
		return errors.Wrapf(err, "can't execute deleting query for film %d", id)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "can't get number affected rows of deleting query for film %d", id)
	}

	if n != 1 {
		return errors.Wrapf(ErrorFilmNotFound, "with id %d", id)
	}

	return nil
}

func (pf *PostgresFilm) GetFilms(params Params) ([]FilmWithActors, error) {
	// По умолчанию ищем в фильмах. Если не задана строка будет поиск всего
	preparedQuery := getFilmsSearchFilm
	if params.SearchField == types.ActorField {
		preparedQuery = getFilmsSearchActor
	}

	if params.SearchString == "*" {
		params.SearchString = ""
	}

	preparedQuery = fmt.Sprintf(preparedQuery, params.OrderField, params.Order)

	tx, err := pf.db.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "can't create transaction for get films")
	}

	// Получаем список всех фильмов
	rows, err := tx.Queryx(preparedQuery, params.SearchString)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't execute get films query")
	}

	films := make([]FilmWithActors, 0)
	filmsIdIndx := make(map[types.Id]uint64)
	filmsId := make([]types.Id, 0)
	i := uint64(0)

	for rows.Next() {
		var film FilmWithActors

		err := rows.Scan(
			&film.ID,
			&film.Name,
			&film.Description,
			&film.DataPublish,
			&film.Rating,
		)

		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "can't scan get films query result")
		}

		film.Actors = make([]Actor, 0)
		films = append(films, film)

		filmsIdIndx[film.ID] = i
		filmsId = append(filmsId, film.ID)

		i++
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't end scan get films query result")
	}

	if len(films) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, errors.Wrap(err, "can't commit transaction for get films")
		}

		return films, nil
	}

	query, args, err := sqlx.In(getFilmsActors, filmsId)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't prepare query to get films actors query")
	}

	query = tx.Rebind(query)

	// Получаем список актёров для каждого фильма
	rows, err = tx.Queryx(query, args...)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't execute get films actors query")
	}

	for rows.Next() {
		var filmId types.Id
		var filmActors Actor

		err := rows.Scan(
			&filmId,
			&filmActors.ID,
			&filmActors.Name,
			&filmActors.Sex,
			&filmActors.Birthday,
		)

		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "can't scan get films actors query result")
		}

		films[filmsIdIndx[filmId]].Actors = append(films[filmsIdIndx[filmId]].Actors, filmActors)
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't end scan get films actors query result")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "can't commit transaction for get films")
	}

	return films, nil
}

const (
	actorIdConflictCode   = "23503"
	actorIdConstraintName = "film_actor_actor_id_fkey"
)

func checkActorConflictError(err error) error {
	var e *pq.Error
	switch {
	case errors.As(err, &e):
		if e.Code == actorIdConflictCode && e.Constraint == actorIdConstraintName {
			return ErrorActorNotFound
		} else {
			return err
		}
	default:
		return err
	}
}
