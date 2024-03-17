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
	"vk_film/internal/repository/film"
	mrf "vk_film/internal/repository/film/mocks"
	"vk_film/pkg/mux"
)

type FilmHandlersSuite struct {
	suite.Suite
	handlers *FilmHandlers
	mockFilm *mrf.FilmRepository
	gmc      *gomock.Controller
}

func (fhs *FilmHandlersSuite) BeforeEach(t provider.T) {
	fhs.gmc = gomock.NewController(t)
	fhs.mockFilm = mrf.NewFilmRepository(fhs.gmc)
	fhs.handlers = NewFilmHandlers(fhs.mockFilm)
}

func (fhs *FilmHandlersSuite) AfterEach(t provider.T) {
	fhs.gmc.Finish()
}

func (fhs *FilmHandlersSuite) TestGetFilmsHandler(t provider.T) {
	t.Title("GetFilms handler of film handlers")
	t.NewStep("Init test data")
	flm := film.FilmWithActors{Film: film.Film{ID: 1, Name: "film", Description: "female"}, Actors: []film.Actor{{}, {}}}
	films := []film.FilmWithActors{flm, flm, flm}

	expectedFilms := response.FromRepositoryFilmsWithActor(films)

	searchString := "search"

	t.WithNewStep("Correct empty params execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().GetFilms(film.Params{
			SearchString: "*",
			SearchField:  types.FilmField,
			OrderField:   types.RatingField,
			Order:        types.DESC,
		}).Return(films, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		var flms []response.Film
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&flms))
		t.Require().EqualValues(expectedFilms, flms)
	})

	t.WithNewStep("Correct all params set with rating, film, desc in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().GetFilms(film.Params{
			SearchString: searchString,
			SearchField:  types.FilmField,
			OrderField:   types.RatingField,
			Order:        types.DESC,
		}).Return(films, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		vals := req.URL.Query()
		vals.Set(OrderFieldKey, string(types.RatingField))
		vals.Set(OrderKey, string(types.DESC))
		vals.Set(SearchFieldKey, string(types.FilmField))
		vals.Set(SearchStringKey, searchString)
		req.URL.RawQuery = vals.Encode()

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		var flms []response.Film
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&flms))
		t.Require().EqualValues(expectedFilms, flms)
	})

	t.WithNewStep("Correct all params set with name, actor, asc in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().GetFilms(film.Params{
			SearchString: searchString,
			SearchField:  types.ActorField,
			OrderField:   types.NameField,
			Order:        types.ASC,
		}).Return(films, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		vals := req.URL.Query()
		vals.Set(OrderFieldKey, string(types.NameField))
		vals.Set(OrderKey, string(types.ASC))
		vals.Set(SearchFieldKey, string(types.ActorField))
		vals.Set(SearchStringKey, searchString)
		req.URL.RawQuery = vals.Encode()

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		var flms []response.Film
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&flms))
		t.Require().EqualValues(expectedFilms, flms)
	})

	t.WithNewStep("Correct all params set with data_publish, film, desc in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().GetFilms(film.Params{
			SearchString: searchString,
			SearchField:  types.FilmField,
			OrderField:   types.DataPublishField,
			Order:        types.DESC,
		}).Return(films, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		vals := req.URL.Query()
		vals.Set(OrderFieldKey, string(types.DataPublishField))
		vals.Set(OrderKey, string(types.DESC))
		vals.Set(SearchFieldKey, string(types.FilmField))
		vals.Set(SearchStringKey, searchString)
		req.URL.RawQuery = vals.Encode()

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusOK, recorder.Code)
		var flms []response.Film
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&flms))
		t.Require().EqualValues(expectedFilms, flms)
	})

	t.WithNewStep("Incorrect search field param value in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		vals := req.URL.Query()
		vals.Set(SearchFieldKey, searchString)
		req.URL.RawQuery = vals.Encode()

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("Incorrect order field param value in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		vals := req.URL.Query()
		vals.Set(OrderFieldKey, searchString)
		req.URL.RawQuery = vals.Encode()

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("Incorrect order param value in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		vals := req.URL.Query()
		vals.Set(OrderKey, searchString)
		req.URL.RawQuery = vals.Encode()

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("Film repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().GetFilms(film.Params{
			SearchString: "*",
			SearchField:  types.FilmField,
			OrderField:   types.RatingField,
			Order:        types.DESC,
		}).Return(films, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, nil)
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.GetFilms(recorder, req, mux.Params{})

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})
}

func (fhs *FilmHandlersSuite) TestUpdateFilmHandler(t provider.T) {
	t.Title("UpdateActor handler of film handlers")
	t.NewStep("Init test data")
	name := "name"
	description := "female"
	data, err := time.Parse("12.02.2202")
	t.Require().NoError(err)
	rating := types.Rating(10)
	actors := []types.Id{1, 2, 3}
	updateFilm := &request.UpdateFilm{
		Name:        &name,
		Description: &description,
		DataPublish: &data,
		Rating:      &rating,
		Actors:      &actors,
	}

	body, err := json.Marshal(updateFilm)
	t.Require().NoError(err)

	updateFilmNil := &request.UpdateFilm{
		Name:        nil,
		Description: nil,
		DataPublish: nil,
		Rating:      nil,
		Actors:      nil,
	}

	bodyNilFilm, err := json.Marshal(updateFilmNil)
	t.Require().NoError(err)

	flm := &film.FilmWithActors{Film: film.Film{ID: 1, Name: "name"}, Actors: []film.Actor{{}, {}}}
	expectedFilm := response.FromRepositoryFilmWithActor(flm)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().UpdateFilm(&film.UpdateFilm{
			ID:           flm.ID,
			Name:         updateFilm.Name,
			Description:  updateFilm.Description,
			DataPublish:  updateFilm.DataPublish,
			Rating:       updateFilm.Rating,
			Actors:       actors,
			UpdateActors: true,
		}).Return(flm, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", flm.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
		var responseFilm response.Film
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseFilm))
		t.Require().EqualValues(*expectedFilm, responseFilm)
	})

	t.WithNewStep("Correct no updates execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().UpdateFilm(&film.UpdateFilm{
			ID:           flm.ID,
			Name:         updateFilmNil.Name,
			Description:  updateFilmNil.Description,
			DataPublish:  updateFilmNil.DataPublish,
			Rating:       updateFilmNil.Rating,
			Actors:       nil,
			UpdateActors: false,
		}).Return(flm, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(bodyNilFilm)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", flm.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
		var responseFilm response.Film
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseFilm))
		t.Require().EqualValues(*expectedFilm, responseFilm)
	})

	t.WithNewStep("Film repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().UpdateFilm(&film.UpdateFilm{
			ID:           flm.ID,
			Name:         updateFilm.Name,
			Description:  updateFilm.Description,
			DataPublish:  updateFilm.DataPublish,
			Rating:       updateFilm.Rating,
			Actors:       actors,
			UpdateActors: true,
		}).Return(flm, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", flm.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Film with no exists actor in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().UpdateFilm(&film.UpdateFilm{
			ID:           flm.ID,
			Name:         updateFilm.Name,
			Description:  updateFilm.Description,
			DataPublish:  updateFilm.DataPublish,
			Rating:       updateFilm.Rating,
			Actors:       actors,
			UpdateActors: true,
		}).Return(flm, film.ErrorActorNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", flm.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusConflict, recorder.Code)
	})

	t.WithNewStep("Film not found error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().UpdateFilm(&film.UpdateFilm{
			ID:           flm.ID,
			Name:         updateFilm.Name,
			Description:  updateFilm.Description,
			DataPublish:  updateFilm.DataPublish,
			Rating:       updateFilm.Rating,
			Actors:       actors,
			UpdateActors: true,
		}).Return(flm, film.ErrorFilmNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", flm.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusNotFound, recorder.Code)
	})

	t.WithNewStep("Body error in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", flm.ID))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("No film id in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("No user permissions in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.UpdateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func (fhs *FilmHandlersSuite) TestDeleteUserHandler(t provider.T) {
	t.Title("DeleteUser handler of film handlers")
	t.NewStep("Init test data")

	filmId := types.Id(1)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().DeleteFilm(filmId).Return(nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", filmId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.DeleteFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusOK, recorder.Code)
	})

	t.WithNewStep("Film repository error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().DeleteFilm(filmId).Return(testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", filmId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.DeleteFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Film repository unknown film execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().DeleteFilm(filmId).Return(film.ErrorFilmNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		req.SetPathValue(FilmIdField, fmt.Sprintf("%d", filmId))
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.DeleteFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusNotFound, recorder.Code)
	})

	t.WithNewStep("Film id not presented in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.DeleteFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusBadRequest, recorder.Code)
	})

	t.WithNewStep("No user permissions in in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(nil, map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.DeleteFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func (fhs *FilmHandlersSuite) TestCreateActorHandler(t provider.T) {
	t.Title("CreateActor handler of film handlers")
	t.NewStep("Init test data")
	createFilm := &request.CreateFilm{
		Name:        "film",
		Description: "female",
		Actors:      []types.Id{1, 2},
	}
	body, err := json.Marshal(createFilm)
	t.Require().NoError(err)

	actors := []types.Id{1, 2}
	createdFilm := &film.Film{ID: 0, Name: "film", Description: "female"}
	flm := &film.FilmWithActors{Film: film.Film{ID: 1, Name: "film", Description: "female"}, Actors: []film.Actor{{}, {}}}
	expectedFilm := response.FromRepositoryFilmWithActor(flm)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().CreateFilm(createdFilm, actors).Return(flm, nil).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.CreateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusCreated, recorder.Code)
		var responseFilm response.Film
		dec := json.NewDecoder(recorder.Body)
		t.Require().NoError(dec.Decode(&responseFilm))
		t.Require().EqualValues(*expectedFilm, responseFilm)
	})

	t.WithNewStep("Film repository error in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().CreateFilm(createdFilm, actors).Return(flm, testError).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.CreateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("Actor no exists in execution", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		fhs.mockFilm.EXPECT().CreateFilm(createdFilm, actors).Return(flm, film.ErrorActorNotFound).Times(1)

		t.NewStep("Init http")
		req, err := initRequest(strings.NewReader(string(body)), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.CreateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusConflict, recorder.Code)
	})

	t.WithNewStep("Body error in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: adminUser})
		t.Require().NoError(err)
		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.CreateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusInternalServerError, recorder.Code)
	})

	t.WithNewStep("No user permissions in execution", func(t provider.StepCtx) {
		t.NewStep("Init http")
		req, err := initRequest(errReader(1), map[types.ContextField]any{middleware.UserField: userUser})
		t.Require().NoError(err)

		recorder := httptest.NewRecorder()

		t.NewStep("Check result")
		fhs.handlers.CreateFilm(recorder, req, *mux.NewParams(req))

		t.Require().Equal(http.StatusForbidden, recorder.Code)
	})
}

func TestRunFilmHandlersSuite(t *testing.T) {
	suite.RunSuite(t, new(FilmHandlersSuite))
}
