package handlers

import (
	"github.com/pkg/errors"
	"net/http"
	"vk_film/internal/delivery/http/v1/model/request"
	"vk_film/internal/delivery/http/v1/model/response"
	"vk_film/internal/delivery/middleware"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/actor"
	"vk_film/pkg/mux"
	"vk_film/pkg/operate"
)

const ActorIdField = "actor_id"

type ActorHandlers struct {
	repository actor.Repository
}

func NewActorHandlers(repository actor.Repository) *ActorHandlers {
	return &ActorHandlers{repository: repository}
}

// CreateActor
//
//	@Summary		Добавление актёра.
//	@Description	Добавляет актёра включая его имя, пол и дату рождения.
//	@Tags			actor
//	@Accept			json
//	@Param			request	body	request.CreateActor	true	"Информация о добавляемом актёре"
//	@Produce		json
//	@Success		201	{object}	response.Actor		"Актёр успешно добавлен в базу"
//	@Failure		400	{object}	operate.ModelError	"В теле запроса ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на создание актёра"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/actor [post]
//	@Security		sessionCookie
func (ah *ActorHandlers) CreateActor(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение значения тела запроса
	var createActor request.CreateActor
	if code, err := parseRequestBody(r.Body, &createActor, request.ValidateCreateActor, l); err != nil {
		operate.SendError(w, err, code, l)
		return
	}

	createdActor, err := ah.repository.CreateActor(&actor.Actor{
		Name:     createActor.Name,
		Sex:      types.Sexes(createActor.Sex),
		Birthday: createActor.Birthday,
	})
	if err != nil {
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't create actor"))
		return
	}

	operate.SendStatus(w, http.StatusCreated, &response.Actor{
		ID:       createdActor.ID,
		Name:     createdActor.Name,
		Sex:      string(createdActor.Sex),
		Birthday: createdActor.Birthday,
	}, l)
}

// DeleteActor
//
//	@Summary		Удаление актёра.
//	@Description	Удаляет информацию об актёре по его id.
//	@Tags			actor
//	@Param			actor_id	path	uint64	true	"Уникальный идентификатор актёра"
//	@Produce		json
//	@Success		200	"Актёр успешно удалён"
//	@Failure		400	{object}	operate.ModelError	"В теле запросе ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на удаление актёра"
//	@Failure		404	{object}	operate.ModelError	"Актёр с указанным id не найден"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/actor/{actor_id} [delete]
//	@Security		sessionCookie
func (ah *ActorHandlers) DeleteActor(w http.ResponseWriter, r *http.Request, params mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение уникального идентификатора
	id, err := params.GetUint64(ActorIdField)
	if err != nil {
		operate.SendError(w, errors.Wrapf(err, "try get actor id"), http.StatusBadRequest, l)
		return
	}

	if err = ah.repository.DeleteActor(types.Id(id)); err != nil {
		if errors.Is(err, actor.ErrorActorNotFound) {
			operate.SendError(w, ErrorActorNotFound, http.StatusNotFound, l)
			return
		}
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't delete actor"))
		return
	}

	operate.SendStatus(w, http.StatusOK, nil, l)
}

// GetActors
//
//	@Summary		Получение списка актёров.
//	@Description	Формирует список всех актёров в системы.
//	@Tags			actor
//	@Produce		json
//	@Success		200	{array}		response.ActorWithFilms	"Список актёров успешно сформирован"
//	@Failure		401	{object}	operate.ModelError		"Пользователь не авторизован"
//	@Failure		500	{object}	operate.ModelError		"Ошибка сервера"
//	@Router			/actor/list [get]
//	@Security		sessionCookie
func (ah *ActorHandlers) GetActors(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	actors, err := ah.repository.GetActors()
	if err != nil {
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't get actors"))
		return
	}

	operate.SendStatus(w, http.StatusOK, response.FromRepositoryActorsWithFilms(actors), l)
}

// UpdateActor
//
//	@Summary		Обновление данных об актёре.
//	@Description	Обновляет данные об актёре. Все переданные поля будут обновлены. Отсутствующие поля будут оставлены без изменений.
//	@Tags			actor
//	@Accept			json
//	@Param			actor_id	path	uint64				true	"Уникальный идентификатор актёра"
//	@Param			request		body	request.UpdateActor	true	"Информация об обновлении"
//	@Produce		json
//	@Success		200	{object}	response.ActorWithFilms	"Актёр успешно обновлен в базе"
//	@Failure		400	{object}	operate.ModelError		"В теле запроса ошибка"
//	@Failure		401	{object}	operate.ModelError		"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError		"У пользователя нет прав на обновление актёра"
//	@Failure		404	{object}	operate.ModelError		"Актёр с указанным id не найден"
//	@Failure		500	{object}	operate.ModelError		"Ошибка сервера"
//	@Router			/actor/{actor_id} [put]
//	@Security		sessionCookie
func (ah *ActorHandlers) UpdateActor(w http.ResponseWriter, r *http.Request, params mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение уникального идентификатора
	id, err := params.GetUint64(ActorIdField)
	if err != nil {
		operate.SendError(w, errors.Wrapf(err, "try get actor id"), http.StatusBadRequest, l)
		return
	}

	// Получение значения тела запроса
	var updateActor request.UpdateActor
	if code, err := parseRequestBody(r.Body, &updateActor, request.ValidateUpdateActor, l); err != nil {
		operate.SendError(w, err, code, l)
		return
	}

	updatedActor, err := ah.repository.UpdateActor(&actor.UpdateActor{
		ID:       types.Id(id),
		Name:     updateActor.Name,
		Sex:      (*types.Sexes)(updateActor.Sex),
		Birthday: updateActor.Birthday,
	})

	if err != nil {
		if errors.Is(err, actor.ErrorActorNotFound) {
			operate.SendError(w, ErrorActorNotFound, http.StatusNotFound, l)
			return
		}
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't update actor"))
		return
	}

	operate.SendStatus(w, http.StatusOK, response.FromRepositoryActorWithFilms(updatedActor), l)
}
