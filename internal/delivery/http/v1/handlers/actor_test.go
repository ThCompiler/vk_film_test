package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"vk_film/internal/delivery/http/v1/model/request"
	"vk_film/internal/delivery/http/v1/model/response"
	"vk_film/internal/delivery/middleware"
	"vk_film/internal/pkg/time"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/actor"
	mra "vk_film/internal/repository/actor/mocks"
	"vk_film/internal/repository/film"
	"vk_film/pkg/mux"
)

type ActorHandlersSuite struct {
	suite.Suite
	handlers  *ActorHandlers
	mockActor *mra.ActorRepository
	gmc       *gomock.Controller
}

func (ahs *ActorHandlersSuite) BeforeEach(t provider.T) {
	ahs.gmc = gomock.NewController(t)
	ahs.mockActor = mra.NewActorRepository(ahs.gmc)
	ahs.handlers = NewActorHandlers(ahs.mockActor)
}

func (ahs *ActorHandlersSuite) AfterEach(t provider.T) {
	ahs.gmc.Finish()
}

func (ahs *ActorHandlersSuite) TestGetActorsHandler(t provider.T) {
	t.Title("GetActors handler of actor handlers")
	t.NewStep("Init test data")
	actr := actor.ActorWithFilms{Actor: actor.Actor{ID: 1}, Films: []film.Film{{}, {}}}
	actors := []actor.ActorWithFilms{actr, actr, actr}
	expectedActors := response.FromRepositoryActorsWithFilms(actors)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().GetActors().Return(actors, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.GetActors(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		var actrs []response.ActorWithFilms
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&actrs))
		t.Require().EqualValues(expectedActors, actrs)
	})

	t.WithNewStep("Actor repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().GetActors().Return(actors, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.GetActors(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})
}

func (ahs *ActorHandlersSuite) TestUpdateActorHandler(t provider.T) {
	t.Title("UpdateActor handler of actor handlers")
	t.NewStep("Init test data")
	name := "name"
	sex := "female"
	data, err := time.Parse("12.02.2202")
	t.Require().NoError(err)
	updateActor := &request.UpdateActor{
		Name:     &name,
		Sex:      &sex,
		Birthday: &data,
	}

	body, err := json.Marshal(updateActor)
	t.Require().NoError(err)

	updateActorNil := &request.UpdateActor{
		Name:     nil,
		Sex:      nil,
		Birthday: nil,
	}

	bodyNilActor, err := json.Marshal(updateActorNil)
	t.Require().NoError(err)

	actr := &actor.ActorWithFilms{Actor: actor.Actor{ID: 1, Name: "name"}, Films: []film.Film{{}, {}}}
	expectedActor := response.FromRepositoryActorWithFilms(actr)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().UpdateActor(&actor.UpdateActor{
			ID:       actr.ID,
			Name:     updateActor.Name,
			Sex:      (*types.Sexes)(updateActor.Sex),
			Birthday: updateActor.Birthday,
		}).Return(actr, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.UpdateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
		var responseActor response.ActorWithFilms
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseActor))
		t.Require().EqualValues(*expectedActor, responseActor)
	})

	t.WithNewStep("Correct no updates execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().UpdateActor(&actor.UpdateActor{
			ID:       actr.ID,
			Name:     updateActorNil.Name,
			Sex:      (*types.Sexes)(updateActorNil.Sex),
			Birthday: updateActorNil.Birthday,
		}).Return(actr, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(bodyNilActor)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.UpdateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
		var responseActor response.ActorWithFilms
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseActor))
		t.Require().EqualValues(*expectedActor, responseActor)
	})

	t.WithNewStep("Actor repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().UpdateActor(&actor.UpdateActor{
			ID:       actr.ID,
			Name:     updateActor.Name,
			Sex:      (*types.Sexes)(updateActor.Sex),
			Birthday: updateActor.Birthday,
		}).Return(actr, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.UpdateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Actor not found error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().UpdateActor(&actor.UpdateActor{
			ID:       actr.ID,
			Name:     updateActor.Name,
			Sex:      (*types.Sexes)(updateActor.Sex),
			Birthday: updateActor.Birthday,
		}).Return(actr, actor.ErrorActorNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.UpdateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusNotFound, recorder.Code)
	})

	t.WithNewStep("Body error in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actr.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.UpdateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("No actor id in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.UpdateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("No user permissions in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.UpdateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func (ahs *ActorHandlersSuite) TestDeleteUserHandler(t provider.T) {
	t.Title("DeleteUser handler of actor handlers")
	t.NewStep("Init test data")

	actorId := types.Id(1)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().DeleteActor(actorId).Return(nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actorId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.DeleteActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
	})

	t.WithNewStep("Actor repository error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().DeleteActor(actorId).Return(testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actorId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.DeleteActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Actor repository unknown actor execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().DeleteActor(actorId).Return(actor.ErrorActorNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(ActorIdField, fmt.Sprintf("%d", actorId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.DeleteActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusNotFound, recorder.Code)
	})

	t.WithNewStep("Actor id not presented in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.DeleteActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("No user permissions in in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.DeleteActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func (ahs *ActorHandlersSuite) TestCreateActorHandler(t provider.T) {
	t.Title("CreateActor handler of actor handlers")
	t.NewStep("Init test data")
	createActor := &request.CreateActor{
		Name: "actor",
		Sex:  "female",
	}
	body, err := json.Marshal(createActor)
	t.Require().NoError(err)

	actr := &actor.Actor{ID: 1, Name: "actor", Sex: "female"}
	createdActor := &actor.Actor{ID: 0, Name: "actor", Sex: "female"}
	expectedActor := &response.Actor{ID: 1, Name: "actor", Sex: "female"}

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().CreateActor(createdActor).Return(actr, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.CreateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusCreated, recorder.Code)
		var responseActor response.Actor
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseActor))
		t.Require().EqualValues(*expectedActor, responseActor)
	})

	t.WithNewStep("Actor repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		ahs.mockActor.EXPECT().CreateActor(createdActor).Return(actr, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.CreateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Body error in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.CreateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("No user permissions in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		ahs.handlers.CreateActor(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func TestRunActorHandlersSuite(t *testing.T) {
	suite.RunSuite(t, new(ActorHandlersSuite))
}
