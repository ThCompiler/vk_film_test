package middleware

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"vk_film/internal/repository/session"
	"vk_film/internal/repository/user"
	"vk_film/internal/usecase/auth/mocks"
	"vk_film/pkg/mux"
)

var testError = errors.New("test error")

type AuthMiddlewareSuite struct {
	suite.Suite
	mockSession *mu.SessionManager
	gmc         *gomock.Controller
}

func (ams *AuthMiddlewareSuite) BeforeEach(t provider.T) {
	ams.gmc = gomock.NewController(t)
	ams.mockSession = mu.NewSessionManager(ams.gmc)
}

func (ams *AuthMiddlewareSuite) AfterEach(t provider.T) {
	ams.gmc.Finish()
}

func (ams *AuthMiddlewareSuite) TestSessionMiddleware(t provider.T) {
	t.Title("Session Middleware")
	t.NewStep("Init test data")
	expectedSessionId := "12314"
	cok := &http.Cookie{}
	cok.Value = expectedSessionId
	cok.Name = string(SessionField)

	expectedUsr := &user.User{
		ID: 1,
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ams.mockSession.EXPECT().GetUserId(expectedSessionId).Return(expectedUsr, nil)

		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)
		reader.AddCookie(cok)

		t.NewStep("Check result")
		CheckSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			usr := r.Context().Value(UserField)
			sessionId := r.Context().Value(SessionField)

			t.Require().NotNil(usr)
			t.Require().NotNil(sessionId)

			var ok bool
			usr, ok = usr.(*user.User)
			t.Require().True(ok)
			t.Require().Equal(expectedUsr, usr)

			sessionId, ok = sessionId.(string)
			t.Require().True(ok)
			t.Require().Equal(expectedSessionId, sessionId)

			w.WriteHeader(http.StatusOK)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
	})

	t.WithNewStep("Session manager error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ams.mockSession.EXPECT().GetUserId(expectedSessionId).Return(expectedUsr, testError)

		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)
		reader.AddCookie(cok)

		t.NewStep("Check result")
		CheckSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			t.Require().True(false)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Session manager no session execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ams.mockSession.EXPECT().GetUserId(expectedSessionId).Return(expectedUsr, session.ErrorNoSession)

		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)
		reader.AddCookie(cok)

		t.NewStep("Check result")
		CheckSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			t.Require().True(false)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusUnauthorized, recorder.Code)
	})

	t.WithNewStep("No session in cookies execute", func(t provider.StepCtx) {
		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)

		t.NewStep("Check result")
		CheckSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			t.Require().True(false)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusUnauthorized, recorder.Code)
	})
}

func (ams *AuthMiddlewareSuite) TestNoSessionMiddleware(t provider.T) {
	t.Title("NoSession Middleware")
	t.NewStep("Init test data")
	expectedSessionId := "12314"
	cok := &http.Cookie{}
	cok.Value = expectedSessionId
	cok.Name = string(SessionField)

	expectedUsr := &user.User{
		ID: 1,
	}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)

		t.NewStep("Check result")
		CheckNoSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			w.WriteHeader(http.StatusOK)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
	})

	t.WithNewStep("Session manager error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ams.mockSession.EXPECT().GetUserId(expectedSessionId).Return(expectedUsr, testError)

		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)
		reader.AddCookie(cok)

		t.NewStep("Check result")
		CheckNoSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			w.WriteHeader(http.StatusOK)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
	})

	t.WithNewStep("Session manager no session execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ams.mockSession.EXPECT().GetUserId(expectedSessionId).Return(expectedUsr, session.ErrorNoSession)

		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)
		reader.AddCookie(cok)

		t.NewStep("Check result")
		CheckNoSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			w.WriteHeader(http.StatusOK)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
	})

	t.WithNewStep("Correct session got in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ams.mockSession.EXPECT().GetUserId(expectedSessionId).Return(expectedUsr, nil)

		t.NewStep("Init http")
		recorder := httptest.NewRecorder()
		reader, err := http.NewRequest(http.MethodPost, "/any", nil)
		t.Require().NoError(err)
		reader.AddCookie(cok)

		t.NewStep("Check result")
		CheckNoSession(ams.mockSession)(func(w http.ResponseWriter, r *http.Request, _ mux.Params) {
			t.Require().True(false)
		})(recorder, reader, mux.Params{})

		t.Require().Equal(http.StatusTeapot, recorder.Code)
	})
}

func TestRunAuthMiddlewareSuite(t *testing.T) {
	suite.RunSuite(t, new(AuthMiddlewareSuite))
}
