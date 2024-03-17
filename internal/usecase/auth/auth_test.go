package auth

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/session"
	mrs "vk_film/internal/repository/session/mocks"
	"vk_film/internal/repository/user"
	mru "vk_film/internal/repository/user/mocks"
)

var testError = errors.New("test error")

type SessionManagerSuite struct {
	suite.Suite
	sessionManager *SessionManager
	mockUser       *mru.UserRepository
	mockSession    *mrs.SessionRepository
	gmc            *gomock.Controller
}

func (sms *SessionManagerSuite) BeforeEach(t provider.T) {
	sms.gmc = gomock.NewController(t)
	sms.mockUser = mru.NewUserRepository(sms.gmc)
	sms.mockSession = mrs.NewSessionRepository(sms.gmc)
	sms.sessionManager = NewSessionManager(sms.mockUser, sms.mockSession)
}

func (sms *SessionManagerSuite) AfterEach(t provider.T) {
	sms.gmc.Finish()
}

func (sms *SessionManagerSuite) TestLoginFunction(t provider.T) {
	t.Title("Login function of sessions manager")
	t.NewStep("Init test data")
	login := "login"
	password := "password"
	encryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	t.Require().NoError(err)
	sessionId := "id"
	userId := types.Id(1)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockUser.EXPECT().GetPasswordByLogin(login).
			Return(&user.LoginUser{ID: userId, Password: string(encryptPassword)}, nil)
		sms.mockSession.EXPECT().Set(gomock.Any(), userId, ExpiredSessionTime).
			Do(
				func(sesId string, _ types.Id, _ time.Duration) {
					sessionId = sesId
				},
			).Return(nil)

		t.NewStep("Check result")
		sesId, err := sms.sessionManager.Login(login, password)
		t.Require().NoError(err)
		t.Require().Equal(sessionId, sesId)
	})

	t.WithNewStep("Incorrect password execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockUser.EXPECT().GetPasswordByLogin(login).
			Return(&user.LoginUser{ID: userId, Password: string(encryptPassword)}, nil)

		t.NewStep("Check result")
		_, err := sms.sessionManager.Login(login, login)
		t.Require().ErrorIs(err, ErrorIncorrectPassword)
	})

	t.WithNewStep("Bcrypt error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockUser.EXPECT().GetPasswordByLogin(login).
			Return(&user.LoginUser{ID: userId, Password: password}, nil)

		t.NewStep("Check result")
		_, err := sms.sessionManager.Login(login, password)
		t.Require().Error(err)
	})

	t.WithNewStep("User repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockUser.EXPECT().GetPasswordByLogin(login).
			Return(&user.LoginUser{ID: userId, Password: password}, testError)

		t.NewStep("Check result")
		_, err := sms.sessionManager.Login(login, password)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Session repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockUser.EXPECT().GetPasswordByLogin(login).
			Return(&user.LoginUser{ID: userId, Password: string(encryptPassword)}, nil)
		sms.mockSession.EXPECT().Set(gomock.Any(), userId, ExpiredSessionTime).Return(testError)

		t.NewStep("Check result")
		_, err := sms.sessionManager.Login(login, password)
		t.Require().ErrorIs(err, testError)
	})
}

func (sms *SessionManagerSuite) TestLogoutFunction(t provider.T) {
	t.Title("Logout function of sessions manager")
	t.NewStep("Init test data")
	sessionId := "id"

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockSession.EXPECT().Del(sessionId).Return(nil)

		t.NewStep("Check result")
		err := sms.sessionManager.Logout(sessionId)
		t.Require().NoError(err)
	})

	t.WithNewStep("Session repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockSession.EXPECT().Del(sessionId).Return(testError)

		t.NewStep("Check result")
		err := sms.sessionManager.Logout(sessionId)
		t.Require().ErrorIs(err, testError)
	})
}

func (sms *SessionManagerSuite) TestGetUserIdFunction(t provider.T) {
	t.Title("GetUserId function of sessions manager")
	t.NewStep("Init test data")
	u := &user.User{
		ID:    1,
		Login: "login",
		Role:  types.USER,
	}
	sessionId := "id"

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockSession.EXPECT().GetUserId(sessionId, ExpiredSessionTime).Return(u.ID, nil)
		sms.mockUser.EXPECT().GetUserById(u.ID).Return(u, nil)

		t.NewStep("Check result")
		usr, err := sms.sessionManager.GetUserId(sessionId)
		t.Require().NoError(err)
		t.Require().EqualValues(u, usr)
	})

	t.WithNewStep("Session repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockSession.EXPECT().GetUserId(sessionId, ExpiredSessionTime).Return(u.ID, testError)

		t.NewStep("Check result")
		_, err := sms.sessionManager.GetUserId(sessionId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("User repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockSession.EXPECT().GetUserId(sessionId, ExpiredSessionTime).Return(u.ID, nil)
		sms.mockUser.EXPECT().GetUserById(u.ID).Return(u, testError)

		t.NewStep("Check result")
		_, err := sms.sessionManager.GetUserId(sessionId)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("User repository user not found in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		sms.mockSession.EXPECT().GetUserId(sessionId, ExpiredSessionTime).Return(u.ID, nil)
		sms.mockUser.EXPECT().GetUserById(u.ID).Return(u, user.ErrorUserNotFound)
		sms.mockSession.EXPECT().Del(sessionId)

		t.NewStep("Check result")
		_, err := sms.sessionManager.GetUserId(sessionId)
		t.Require().ErrorIs(err, session.ErrorNoSession)
	})
}

func TestRunSessionManagerSuite(t *testing.T) {
	suite.RunSuite(t, new(SessionManagerSuite))
}
