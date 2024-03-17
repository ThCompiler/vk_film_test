package user

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"vk_film/internal/pkg/types"
)

const (
	createQuery = `
		WITH sel AS (
				SELECT id, login, role
				FROM users
				WHERE login = $1 LIMIT 1
		), ins as (
			INSERT INTO users (login, password, role)
				SELECT $1, $2, $3
			    WHERE not exists (select 1 from sel)
			RETURNING id, login, role
		)
		SELECT id, login, role, 0
		FROM ins
		UNION ALL
		SELECT id, login, role, 1
		FROM sel
	`

	updateUser = `
		UPDATE users SET role = $2 WHERE id = $1 RETURNING id, login, role
	`

	deleteUser = `
		DELETE FROM users WHERE id = $1
	`

	getUsers = `
		SELECT id, login, role FROM users
	`

	getPasswordByLogin = `
		SELECT id, password FROM users WHERE login = $1
	`

	getUserById = `
		SELECT id, login, role FROM users WHERE id = $1
	`
)

type PostgresUser struct {
	db *sqlx.DB
}

func NewPostgresUser(db *sqlx.DB) *PostgresUser {
	return &PostgresUser{
		db: db,
	}
}

func (pu *PostgresUser) CreateUser(user *User) (*User, error) {
	newUser := &User{}
	exists := false
	if err := pu.db.QueryRowx(createQuery, user.Login, user.Password, user.Role).
		Scan(
			&newUser.ID,
			&newUser.Login,
			&newUser.Role,
			&exists,
		); err != nil {
		return nil, errors.Wrap(err, "can't create user")
	}

	if exists {
		return newUser, ErrorLoginAlreadyExists
	}

	return newUser, nil
}

func (pu *PostgresUser) UpdateUserRole(user *User) (*User, error) {
	updatedUser := &User{}

	if err := pu.db.QueryRowx(updateUser, user.ID, user.Role).
		Scan(
			&updatedUser.ID,
			&updatedUser.Login,
			&updatedUser.Role,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorUserNotFound
		}
		return nil, errors.Wrapf(err, "can't update user with id %d", user.ID)
	}

	return updatedUser, nil
}

func (pu *PostgresUser) DeleteUser(id types.Id) error {
	res, err := pu.db.Exec(deleteUser, id)
	if err != nil {
		return errors.Wrapf(err, "can't execute deleting query for user %d", id)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "can't get number affected rows of deleting query for user %d", id)
	}

	if n != 1 {
		return errors.Wrapf(ErrorUserNotFound, "with id %d", id)
	}

	return nil
}

func (pu *PostgresUser) GetPasswordByLogin(login string) (*LoginUser, error) {
	lu := &LoginUser{}

	if err := pu.db.QueryRowx(getPasswordByLogin, login).
		Scan(
			&lu.ID,
			&lu.Password,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorUserNotFound
		}
		return nil, errors.Wrapf(err, "can't found user by login %s", login)
	}

	return lu, nil
}

func (pu *PostgresUser) GetUserById(id types.Id) (*User, error) {
	foundedUser := &User{}

	if err := pu.db.QueryRowx(getUserById, id).
		Scan(
			&foundedUser.ID,
			&foundedUser.Login,
			&foundedUser.Role,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorUserNotFound
		}
		return nil, errors.Wrapf(err, "can't found user by id %d", id)
	}

	return foundedUser, nil
}

func (pu *PostgresUser) GetUsers() ([]User, error) {
	rows, err := pu.db.Queryx(getUsers)
	if err != nil {
		return nil, errors.Wrap(err, "can't execute get users query")
	}

	users := make([]User, 0)

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.Role,
		)

		if err != nil {
			return nil, errors.Wrap(err, "can't scan get users query result")
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "can't end scan get users query result")
	}

	return users, nil
}
