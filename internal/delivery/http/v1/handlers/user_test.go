package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"vk_film/internal/delivery/http/v1/model/request"
	"vk_film/internal/delivery/http/v1/model/response"
	"vk_film/internal/delivery/middleware"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/user"
	mru "vk_film/internal/repository/user/mocks"
	"vk_film/internal/usecase/auth"
	mua "vk_film/internal/usecase/auth/mocks"
	"vk_film/pkg/mux"
)

type UserHandlersSuite struct {
	suite.Suite
	handlers *UserHandlers
	mockUser *mru.UserRepository
	mockAuth *mua.SessionManager
	gmc      *gomock.Controller
}

func (uhs *UserHandlersSuite) BeforeEach(t provider.T) {
	uhs.gmc = gomock.NewController(t)
	uhs.mockUser = mru.NewUserRepository(uhs.gmc)
	uhs.mockAuth = mua.NewSessionManager(uhs.gmc)
	uhs.handlers = NewUserHandlers(uhs.mockUser, uhs.mockAuth)
}

func (uhs *UserHandlersSuite) AfterEach(t provider.T) {
	uhs.gmc.Finish()
}

func (uhs *UserHandlersSuite) TestLoginHandler(t provider.T) {
	t.Title("Login handler of user handlers")
	t.NewStep("Init test data")
	login := "login"
	password := "password"
	sessionId := "id"
	body := "{ \"login\": \"login\", \"password\": \"password\" }"

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockAuth.EXPECT().Login(login, password).
			Return(sessionId, nil).Times(1)

		t.NewStep("Init http")

		req, err := initRequest(strings.NewReader(body), nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.Login(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		cks := recorder.Result().Cookies()
		i := slices.IndexFunc(cks,
			func(ck *http.Cookie) bool { return ck != nil && ck.Name == string(middleware.SessionField) },
		)
		t.Require().NotEqual(-1, i)
		t.Require().Equal(sessionId, cks[i].Value)
	})

	t.WithNewStep("Incorrect Body execute", func(t provider.StepCtx) {
		t.NewStep("Init http")

		req, err := initRequest(errReader(1), nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.Login(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Session manager error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockAuth.EXPECT().Login(login, password).
			Return(sessionId, testError).Times(1)

		t.NewStep("Init http")

		req, err := initRequest(strings.NewReader(body), nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.Login(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Incorrect password in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockAuth.EXPECT().Login(login, password).
			Return(sessionId, auth.ErrorIncorrectPassword).Times(1)

		t.NewStep("Init http")

		req, err := initRequest(strings.NewReader(body), nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.Login(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusConflict, recorder.Code)
	})

	t.WithNewStep("Incorrect login in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockAuth.EXPECT().Login(login, password).
			Return(sessionId, user.ErrorUserNotFound).Times(1)

		t.NewStep("Init http")

		req, err := initRequest(strings.NewReader(body), nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.Login(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusConflict, recorder.Code)
	})
}

func (uhs *UserHandlersSuite) TestLogoutHandler(t provider.T) {
	t.Title("Logout handler of user handlers")
	t.NewStep("Init test data")
	sessionId := "id"

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockAuth.EXPECT().Logout(sessionId).Return(nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.SessionField: sessionId})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.Logout(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		cks := recorder.Result().Cookies()
		i := slices.IndexFunc(cks,
			func(ck *http.Cookie) bool { return ck != nil && ck.Name == string(middleware.SessionField) },
		)
		t.Require().NotEqual(-1, i)
		t.Require().Equal("", cks[i].Value)
	})

	t.WithNewStep("Session manager error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockAuth.EXPECT().Logout(sessionId).Return(testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.SessionField: sessionId})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.Logout(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		cks := recorder.Result().Cookies()
		i := slices.IndexFunc(cks,
			func(ck *http.Cookie) bool { return ck != nil && ck.Name == string(middleware.SessionField) },
		)
		t.Require().NotEqual(-1, i)
		t.Require().Equal("", cks[i].Value)
	})
}

func (uhs *UserHandlersSuite) TestGetUsersHandler(t provider.T) {
	t.Title("GetUsers handler of user handlers")
	t.NewStep("Init test data")
	users := []user.User{{ID: 1}, {ID: 2}, {ID: 3}}
	expectedUsers := []response.User{{ID: 1}, {ID: 2}, {ID: 3}}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().GetUsers().Return(users, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.GetUsers(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		var usrs []response.User
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&usrs))
		t.Require().EqualValues(expectedUsers, usrs)
	})

	t.WithNewStep("User repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().GetUsers().Return(users, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.GetUsers(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})
}

func (uhs *UserHandlersSuite) TestUpdateUserRoleHandler(t provider.T) {
	t.Title("UpdateUserRole handler of user handlers")
	t.NewStep("Init test data")
	updatedRole := &request.UpdateRole{
		Role: string(types.USER),
	}
	body, err := json.Marshal(updatedRole)
	t.Require().NoError(err)

	usr := &user.User{ID: 1, Role: types.USER}
	expectedUser := &response.User{ID: 1, Role: string(types.USER)}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().UpdateUserRole(usr).Return(usr, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(UserIdField, fmt.Sprintf("%d", usr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.UpdateUserRole(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
		var responseUser response.User
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseUser))
		t.Require().EqualValues(*expectedUser, responseUser)
	})

	t.WithNewStep("User repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().UpdateUserRole(usr).Return(usr, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(UserIdField, fmt.Sprintf("%d", usr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.UpdateUserRole(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("User not found error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().UpdateUserRole(usr).Return(usr, user.ErrorUserNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(UserIdField, fmt.Sprintf("%d", usr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.UpdateUserRole(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusNotFound, recorder.Code)
	})

	t.WithNewStep("Body error in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(UserIdField, fmt.Sprintf("%d", usr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.UpdateUserRole(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("No user id in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.UpdateUserRole(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("No user permissions in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.UpdateUserRole(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func (uhs *UserHandlersSuite) TestDeleteUserHandler(t provider.T) {
	t.Title("DeleteUser handler of user handlers")
	t.NewStep("Init test data")

	userId := types.Id(1)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().DeleteUser(userId).Return(nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(UserIdField, fmt.Sprintf("%d", userId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.DeleteUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
	})

	t.WithNewStep("User repository error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().DeleteUser(userId).Return(testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(UserIdField, fmt.Sprintf("%d", userId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.DeleteUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("User repository unknown user execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().DeleteUser(userId).Return(user.ErrorUserNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(UserIdField, fmt.Sprintf("%d", userId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.DeleteUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusNotFound, recorder.Code)
	})

	t.WithNewStep("User id not presented in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.DeleteUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("No user permissions in in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.DeleteUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

type CreateUserMather user.User

func (um *CreateUserMather) Matches(x any) bool {
	if usr, ok := x.(*user.User); ok {
		return usr.Login == um.Login && usr.Role == um.Role &&
			bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(um.Password)) == nil
	}
	return false
}

func (um *CreateUserMather) String() string {
	return fmt.Sprintf("is equal to %v (%T)", *um, um)
}

func (uhs *UserHandlersSuite) TestCreateUserHandler(t provider.T) {
	t.Title("CreateUser handler of user handlers")
	t.NewStep("Init test data")
	createUser := &request.CreateUser{
		Role:     string(types.USER),
		Password: "1",
	}
	body, err := json.Marshal(createUser)
	t.Require().NoError(err)

	createUserWithoutRole := &request.CreateUser{
		Password: "1",
	}
	bodyWithOutRole, err := json.Marshal(createUserWithoutRole)
	t.Require().NoError(err)

	usr := &user.User{ID: 1, Role: types.USER, Password: "1"}
	expectedUser := &response.User{ID: 1, Role: string(types.USER)}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().CreateUser((*CreateUserMather)(usr)).Return(usr, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.CreateUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusCreated, recorder.Code)
		var responseUser response.User
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseUser))
		t.Require().EqualValues(*expectedUser, responseUser)
	})

	t.WithNewStep("Correct execute default role", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().CreateUser((*CreateUserMather)(usr)).Return(usr, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(bodyWithOutRole)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.CreateUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusCreated, recorder.Code)
		var responseUser response.User
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseUser))
		t.Require().EqualValues(*expectedUser, responseUser)
	})

	t.WithNewStep("User repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().CreateUser((*CreateUserMather)(usr)).Return(usr, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.CreateUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("User already exists error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		uhs.mockUser.EXPECT().CreateUser((*CreateUserMather)(usr)).Return(usr, user.ErrorLoginAlreadyExists).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.CreateUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusConflict, recorder.Code)
	})

	t.WithNewStep("Body error in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.CreateUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("No user permissions in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		uhs.handlers.CreateUser(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func TestRunUserHandlersSuite(t *testing.T) {
	suite.RunSuite(t, new(UserHandlersSuite))
}
