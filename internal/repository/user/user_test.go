package user

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"testing"
	"vk_film/internal/pkg/types"
)

var testError = errors.New("test error")

type UserRepositorySuite struct {
	suite.Suite
	userRepository *PostgresUser
	mock           sqlxmock.Sqlmock
}

func (urs *UserRepositorySuite) BeforeEach(t provider.T) {
	db, mock, err := sqlxmock.Newx(sqlxmock.QueryMatcherOption(sqlxmock.QueryMatcherEqual))
	t.Require().NoError(err)
	urs.userRepository = NewPostgresUser(db)
	urs.mock = mock
}

func (urs *UserRepositorySuite) AfterEach(t provider.T) {
	t.Require().NoError(urs.mock.ExpectationsWereMet())
}

func (urs *UserRepositorySuite) TestCreateFunction(t provider.T) {
	t.Title("CreateUser function of User repository")
	t.NewStep("Init test data")
	user := &User{
		ID:       1,
		Login:    "actor",
		Password: "password",
		Role:     types.USER,
	}

	userColumns := []string{
		"id", "login", "role", "exists",
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(createQuery).
			WithArgs(user.Login, user.Password, user.Role).
			WillReturnRows(sqlxmock.NewRows(userColumns).
				AddRow(user.ID, user.Login, user.Role, 0),
			)

		t.NewStep("Check result")
		usr, err := urs.userRepository.CreateUser(user)
		t.Require().NoError(err)
		t.Require().EqualValues(&User{
			ID:    user.ID,
			Login: user.Login,
			Role:  user.Role,
		}, usr)
	})

	t.WithNewStep("Error user already exists in create Query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(createQuery).
			WithArgs(user.Login, user.Password, user.Role).
			WillReturnRows(sqlxmock.NewRows(userColumns).
				AddRow(user.ID, user.Login, user.Role, 1),
			)

		t.NewStep("Check result")
		_, err := urs.userRepository.CreateUser(user)
		t.Require().ErrorIs(err, ErrorLoginAlreadyExists)
	})

	t.WithNewStep("Postgres error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(createQuery).
			WithArgs(user.Login, user.Password, user.Role).
			WillReturnError(testError)

		t.NewStep("Check result")
		_, err := urs.userRepository.CreateUser(user)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Empty result of execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(createQuery).
			WithArgs(user.Login, user.Password, user.Role).
			WillReturnRows(sqlxmock.NewRows(userColumns))

		t.NewStep("Check result")
		_, err := urs.userRepository.CreateUser(user)
		t.Require().Error(err)
	})
}

func (urs *UserRepositorySuite) TestDeleteFunction(t provider.T) {
	t.Title("DeleteUser function of User repository")
	t.NewStep("Init test data")
	user := &User{
		ID:       1,
		Login:    "actor",
		Password: "password",
		Role:     types.USER,
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectExec(deleteUser).
			WithArgs(user.ID).
			WillReturnResult(sqlxmock.NewResult(0, 1))

		t.NewStep("Check result")
		err := urs.userRepository.DeleteUser(user.ID)
		t.Require().NoError(err)
	})

	t.WithNewStep("Postgres error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectExec(deleteUser).
			WithArgs(user.ID).
			WillReturnError(testError)

		t.NewStep("Check result")
		err := urs.userRepository.DeleteUser(user.ID)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Row affected error of execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectExec(deleteUser).
			WithArgs(user.ID).
			WillReturnResult(sqlxmock.NewErrorResult(testError))

		t.NewStep("Check result")
		err := urs.userRepository.DeleteUser(user.ID)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Error not found user in execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectExec(deleteUser).
			WithArgs(user.ID).
			WillReturnResult(sqlxmock.NewResult(2, 0))

		t.NewStep("Check result")
		err := urs.userRepository.DeleteUser(user.ID)
		t.Require().ErrorIs(err, ErrorUserNotFound)
	})
}

func (urs *UserRepositorySuite) TestGetFunction(t provider.T) {
	t.Title("GetUsers function of User repository")
	t.NewStep("Init test data")
	user := &User{
		ID:    1,
		Login: "actor",
		Role:  types.USER,
	}

	userColumns := []string{
		"id", "login", "role",
	}

	usersRows := func() *sqlxmock.Rows {
		return sqlxmock.NewRows(userColumns).
			AddRow(user.ID, user.Login, user.Role).
			AddRow(user.ID, user.Login, user.Role).
			AddRow(user.ID, user.Login, user.Role)
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUsers).WillReturnRows(usersRows())

		t.NewStep("Check result")
		users, err := urs.userRepository.GetUsers()
		t.Require().NoError(err)
		t.Require().EqualValues([]User{*user, *user, *user}, users)
	})

	t.WithNewStep("Empty list in execute result", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUsers).WillReturnRows(sqlxmock.NewRows(userColumns))

		t.NewStep("Check result")
		users, err := urs.userRepository.GetUsers()
		t.Require().NoError(err)
		t.Require().EqualValues([]User{}, users)
	})

	t.WithNewStep("Postgres error on execute query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUsers).WillReturnError(testError)

		t.NewStep("Check result")
		_, err := urs.userRepository.GetUsers()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Rows error on getUsers query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUsers).WillReturnRows(usersRows().RowError(1, testError))

		t.NewStep("Check result")
		_, err := urs.userRepository.GetUsers()
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Incorrect field in row of getUsers query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUsers).WillReturnRows(usersRows().AddRow(1, 1, 1))

		t.NewStep("Check result")
		_, err := urs.userRepository.GetUsers()
		t.Require().Error(err)
	})

	t.WithNewStep("Rows close error on getUsers query", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUsers).WillReturnRows(usersRows().CloseError(testError))

		t.NewStep("Check result")
		_, err := urs.userRepository.GetUsers()
		t.Require().ErrorIs(err, testError)
	})
}

func (urs *UserRepositorySuite) TestUpdateRoleFunction(t provider.T) {
	t.Title("UpdateUserRole function of User repository")
	t.NewStep("Init test data")
	user := &User{
		ID:    1,
		Login: "actor",
		Role:  types.USER,
	}

	userColumns := []string{
		"id", "login", "role",
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(updateUser).
			WithArgs(user.ID, user.Role).
			WillReturnRows(
				sqlxmock.NewRows(userColumns).
					AddRow(user.ID, user.Login, user.Role),
			)

		t.NewStep("Check result")
		usr, err := urs.userRepository.UpdateUserRole(user)
		t.Require().NoError(err)
		t.Require().EqualValues(user, usr)
	})

	t.WithNewStep("Postgres error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(updateUser).
			WithArgs(user.ID, user.Role).
			WillReturnError(testError)

		t.NewStep("Check result")
		_, err := urs.userRepository.UpdateUserRole(user)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("User not found to update on execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(updateUser).
			WithArgs(user.ID, user.Role).
			WillReturnRows(sqlxmock.NewRows(userColumns))

		t.NewStep("Check result")
		_, err := urs.userRepository.UpdateUserRole(user)
		t.Require().ErrorIs(err, ErrorUserNotFound)
	})
}

func (urs *UserRepositorySuite) TestGetUserByIdFunction(t provider.T) {
	t.Title("GetUserById function of User repository")
	t.NewStep("Init test data")
	user := &User{
		ID:    1,
		Login: "actor",
		Role:  types.USER,
	}

	userColumns := []string{
		"id", "login", "role",
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUserById).
			WithArgs(user.ID).
			WillReturnRows(
				sqlxmock.NewRows(userColumns).
					AddRow(user.ID, user.Login, user.Role),
			)

		t.NewStep("Check result")
		usr, err := urs.userRepository.GetUserById(user.ID)
		t.Require().NoError(err)
		t.Require().EqualValues(user, usr)
	})

	t.WithNewStep("Postgres error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUserById).
			WithArgs(user.ID).
			WillReturnError(testError)

		t.NewStep("Check result")
		_, err := urs.userRepository.GetUserById(user.ID)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("User not found to get on execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getUserById).
			WithArgs(user.ID).
			WillReturnRows(sqlxmock.NewRows(userColumns))

		t.NewStep("Check result")
		_, err := urs.userRepository.GetUserById(user.ID)
		t.Require().ErrorIs(err, ErrorUserNotFound)
	})
}

func (urs *UserRepositorySuite) TestGetPasswordByLoginFunction(t provider.T) {
	t.Title("GetPasswordByLogin function of User repository")
	t.NewStep("Init test data")
	user := &User{
		ID:    1,
		Login: "actor",
		Role:  types.USER,
	}

	userColumns := []string{
		"id", "password",
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getPasswordByLogin).
			WithArgs(user.Login).
			WillReturnRows(
				sqlxmock.NewRows(userColumns).
					AddRow(user.ID, user.Password),
			)

		t.NewStep("Check result")
		usr, err := urs.userRepository.GetPasswordByLogin(user.Login)
		t.Require().NoError(err)
		t.Require().EqualValues(&LoginUser{
			ID:       user.ID,
			Password: user.Password,
		}, usr)
	})

	t.WithNewStep("Postgres error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getPasswordByLogin).
			WithArgs(user.Login).
			WillReturnError(testError)

		t.NewStep("Check result")
		_, err := urs.userRepository.GetPasswordByLogin(user.Login)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("User not found to get on execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		urs.mock.ExpectQuery(getPasswordByLogin).
			WithArgs(user.Login).
			WillReturnRows(sqlxmock.NewRows(userColumns))

		t.NewStep("Check result")
		_, err := urs.userRepository.GetPasswordByLogin(user.Login)
		t.Require().ErrorIs(err, ErrorUserNotFound)
	})
}

func TestRunUserRepositorySuite(t *testing.T) {
	suite.RunSuite(t, new(UserRepositorySuite))
}
