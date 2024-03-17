package actor

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/film"
)

const (
	createQuery = `
		INSERT INTO actors (name, sex, birthday)
		VALUES ($1, $2, $3)
		RETURNING id, name, sex, birthday
	`

	deleteActor = `
		DELETE FROM actors WHERE id = $1
	`

	updateActors = `
		UPDATE actors SET name = upd_actor.upd_name, sex = upd_actor.upd_sex, birthday = upd_actor.upd_birthday 
			FROM (
				SELECT COALESCE($2, actors.name) as upd_name, 
					   COALESCE($3, actors.sex) as upd_sex, 
					   COALESCE($4, actors.birthday) as upd_birthday 
				FROM actors WHERE id = $1
			) as upd_actor
			WHERE id = $1
			RETURNING id, name, sex, birthday
	`

	getActorFilms = `
		SELECT films.id, films.name, films.description, films.publish_date, films.rating FROM film_actor
			JOIN films on (films.id = film_actor.film_id)
		 	WHERE film_actor.actor_id = $1
	`

	getActors = `
		SELECT id, name, sex, birthday FROM actors
	`

	getActorsFilms = `
		SELECT actors.id, films.id, films.name, films.description, films.publish_date, films.rating FROM actors 
			JOIN film_actor on (actors.id = film_actor.actor_id)
			JOIN films on (films.id = film_actor.film_id)
	`
)

type PostgresActor struct {
	db *sqlx.DB
}

func NewPostgresActor(db *sqlx.DB) *PostgresActor {
	return &PostgresActor{
		db: db,
	}
}

func (pa *PostgresActor) CreateActor(actor *Actor) (*Actor, error) {
	newActor := &Actor{}

	if err := pa.db.QueryRowx(createQuery, actor.Name, actor.Sex, &actor.Birthday).
		Scan(
			&newActor.ID,
			&newActor.Name,
			&newActor.Sex,
			&newActor.Birthday,
		); err != nil {
		return nil, errors.Wrap(err, "can't create actor")
	}

	return newActor, nil
}

func getNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{Valid: true, String: *value}
}

func (pa *PostgresActor) UpdateActor(actor *UpdateActor) (*ActorWithFilms, error) {
	name := getNullString(actor.Name)
	sex := getNullString((*string)(actor.Sex))

	birthday := sql.NullTime{Valid: false}
	if actor.Birthday != nil {
		birthday = sql.NullTime{Valid: true, Time: actor.Birthday.Time}
	}

	tx, err := pa.db.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "can't create transaction for update actor")
	}

	updatedActor := &ActorWithFilms{}

	if err := tx.QueryRowx(updateActors, actor.ID, name, sex, birthday).
		Scan(
			&updatedActor.ID,
			&updatedActor.Name,
			&updatedActor.Sex,
			&updatedActor.Birthday,
		); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorActorNotFound
		}
		return nil, errors.Wrapf(err, "can't update actor with id %d", actor.ID)
	}

	// Получаем список фильмов для автора
	rows, err := tx.Queryx(getActorFilms, actor.ID)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't execute get updated actor films query")
	}

	for rows.Next() {
		var actorFilms film.Film

		err := rows.Scan(
			&actorFilms.ID,
			&actorFilms.Name,
			&actorFilms.Description,
			&actorFilms.DataPublish,
			&actorFilms.Rating,
		)

		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "can't scan get updated actor films query result")
		}

		updatedActor.Films = append(updatedActor.Films, actorFilms)
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't end scan get updated actor films query result")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "can't commit transaction for update actor")
	}

	return updatedActor, nil
}

func (pa *PostgresActor) DeleteActor(id types.Id) error {
	res, err := pa.db.Exec(deleteActor, id)
	if err != nil {
		return errors.Wrapf(err, "can't execute deleting query for actor %d", id)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "can't get number affected rows of deleting query for actor %d", id)
	}

	if n < 1 {
		return errors.Wrapf(ErrorActorNotFound, "with id %d", id)
	}

	return nil
}

func (pa *PostgresActor) GetActors() ([]ActorWithFilms, error) {
	tx, err := pa.db.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "can't create transaction for get actors")
	}

	// Получаем список всех авторов
	rows, err := tx.Queryx(getActors)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't execute get actors query")
	}

	actors := make([]ActorWithFilms, 0)
	actorsId := make(map[types.Id]uint64)
	i := uint64(0)

	for rows.Next() {
		var actor ActorWithFilms

		err := rows.Scan(
			&actor.ID,
			&actor.Name,
			&actor.Sex,
			&actor.Birthday,
		)

		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "can't scan get actors query result")
		}

		actor.Films = make([]film.Film, 0)

		actors = append(actors, actor)
		actorsId[actor.ID] = i
		i++
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't end scan get actors query result")
	}

	// Получаем список фильмов для каждого автора
	rows, err = tx.Queryx(getActorsFilms)
	if err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't execute get actors films query")
	}

	for rows.Next() {
		var actorId types.Id
		var actorFilms film.Film

		err := rows.Scan(
			&actorId,
			&actorFilms.ID,
			&actorFilms.Name,
			&actorFilms.Description,
			&actorFilms.DataPublish,
			&actorFilms.Rating,
		)

		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "can't scan get actors films query result")
		}

		actors[actorsId[actorId]].Films = append(actors[actorsId[actorId]].Films, actorFilms)
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, "can't end scan get actors films query result")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "can't commit transaction for get actors")
	}

	return actors, nil
}
