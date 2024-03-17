package handlers

import (
	"github.com/pkg/errors"
	"net/http"
	"vk_film/internal/delivery/http/v1/model/request"
	"vk_film/internal/delivery/http/v1/model/response"
	"vk_film/internal/delivery/middleware"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/film"
	"vk_film/pkg/mux"
	"vk_film/pkg/operate"
)

const (
	OrderKey        = "sort_order"
	OrderFieldKey   = "sort_by"
	SearchStringKey = "search_string"
	SearchFieldKey  = "search_by"

	FilmIdField = "film_id"
)

type FilmHandlers struct {
	repository film.Repository
}

func NewFilmHandlers(repository film.Repository) *FilmHandlers {
	return &FilmHandlers{repository: repository}
}

// CreateFilm
//
//	@Summary		Добавление фильма.
//	@Description	Добавляет фильм включая его название, описание, рейтинг, дату публикации и список игравших в нём актёров.
//	@Tags			film
//	@Accept			json
//	@Param			request	body	request.CreateFilm	true	"Информация о добавляемом фильме"
//	@Produce		json
//	@Success		201	{object}	response.Film		"Фильм успешно добавлен в базу"
//	@Failure		400	{object}	operate.ModelError	"В теле запроса ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на создание фильма"
//	@Failure		409	{object}	operate.ModelError	"Актёр фильма не найден"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/film [post]
//	@Security		sessionCookie
func (fh *FilmHandlers) CreateFilm(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение значения тела запроса
	var createFilm request.CreateFilm
	if code, err := parseRequestBody(r.Body, &createFilm, request.ValidateCreateFilm, l); err != nil {
		operate.SendError(w, err, code, l)
		return
	}

	createdFilm, err := fh.repository.CreateFilm(&film.Film{
		Name:        createFilm.Name,
		Description: createFilm.Description,
		DataPublish: createFilm.DataPublish,
		Rating:      createFilm.Rating,
	}, createFilm.Actors)
	if err != nil {
		if errors.Is(err, film.ErrorActorNotFound) {
			operate.SendError(w, ErrorActorNotFound, http.StatusConflict, l)
			l.Info(err)
			return
		}
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't create film"))
		return
	}

	operate.SendStatus(w, http.StatusCreated, response.FromRepositoryFilmWithActor(createdFilm), l)
}

// DeleteFilm
//
//	@Summary		Удаление фильма.
//	@Description	Удаляет информацию о фильме из системы по его id.
//	@Tags			film
//	@Param			film_id	path	uint64	true	"Уникальный идентификатор фильма"
//	@Produce		json
//	@Success		200	"Фильм успешно удалён"
//	@Failure		400	{object}	operate.ModelError	"В теле запросе ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на удаление фильма"
//	@Failure		404	{object}	operate.ModelError	"Фильм с указанным id не найден"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/film/{film_id} [delete]
//	@Security		sessionCookie
func (fh *FilmHandlers) DeleteFilm(w http.ResponseWriter, r *http.Request, params mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение уникального идентификатора
	id, err := params.GetUint64(FilmIdField)
	if err != nil {
		operate.SendError(w, errors.Wrapf(err, "try get film id"), http.StatusBadRequest, l)
		return
	}

	if err = fh.repository.DeleteFilm(types.Id(id)); err != nil {
		if errors.Is(err, film.ErrorFilmNotFound) {
			operate.SendError(w, ErrorFilmNotFound, http.StatusNotFound, l)
			return
		}
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't delete film"))
		return
	}

	operate.SendStatus(w, http.StatusOK, nil, l)
}

// GetFilms
//
//	@Summary		Получение списка фильмов.
//	@Description	Позволяет получить список фильмом отсортированный по определённому полю. А также можно делать поиска списка фильма по имени актёра или названии фильма. Если параметры "search_by" и "search_string" не указаны, поиск не производится.
//	@Tags			film
//	@Param			sort_order		query	string	false	"Порядок сортировки. Возможна сортировка по возрастанию 'asc' или по убыванию 'desc'."											Enums(DESC, ASC)					default(DESC)
//	@Param			sort_by			query	string	false	"Параметр сортировки. Возможна сортировка по рейтингу 'rating', имени 'name' и дате публикации 'publish_date'."																			Enums(rating, name, publish_date)	default(rating)
//	@Param			search_by		query	string	false	"Параметр поиска. Возможен поиск по фрагменту имени актёра 'actor' или фрагменту названия фильма 'film'. Обязателен при указании параметра 'search_name'."	Enums(actor, film)
//	@Param			search_string	query	string	false	"Фргамнет, по которому осуществляется поиск. Обязателен при указании параметра 'search_by'"
//	@Produce		json
//	@Success		200	{array}		response.Film		"Список фильмом успешно сформирован"
//	@Failure		400	{object}	operate.ModelError	"В теле запросе ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/film/list [get]
//	@Security		sessionCookie
func (fh *FilmHandlers) GetFilms(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	// Формируем параметры получения
	getParams := film.Params{
		SearchString: "*",
		SearchField:  types.FilmField,
		OrderField:   types.RatingField,
		Order:        types.DESC,
	}

	values := r.URL.Query()

	if values.Has(SearchStringKey) {
		getParams.SearchString = values.Get(SearchStringKey)
	}

	if values.Has(SearchFieldKey) {
		field := types.SearchField(values.Get(SearchFieldKey))
		if field != types.FilmField && field != types.ActorField {
			operate.SendError(w, ErrorIncorrectQueryParam, http.StatusBadRequest, l)
			l.Warn(errors.Wrapf(ErrorIncorrectQueryParam, "with field %s and value %s", SearchFieldKey, field))
			return
		}

		getParams.SearchField = field
	}

	if values.Has(OrderKey) {
		field := types.Order(values.Get(OrderKey))
		if field != types.DESC && field != types.ASC {
			operate.SendError(w, ErrorIncorrectQueryParam, http.StatusBadRequest, l)
			l.Warn(errors.Wrapf(ErrorIncorrectQueryParam, "with field %s and value %s", OrderKey, field))
			return
		}

		getParams.Order = field
	}

	if values.Has(OrderFieldKey) {
		field := types.OrderField(values.Get(OrderFieldKey))
		if field != types.RatingField && field != types.NameField && field != types.DataPublishField {
			operate.SendError(w, ErrorIncorrectQueryParam, http.StatusBadRequest, l)
			l.Warn(errors.Wrapf(ErrorIncorrectQueryParam, "with field %s and value %s", OrderFieldKey, field))
			return
		}

		getParams.OrderField = field
	}

	films, err := fh.repository.GetFilms(getParams)
	if err != nil {
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't create film"))
		return
	}

	operate.SendStatus(w, http.StatusOK, response.FromRepositoryFilmsWithActor(films), l)
}

// UpdateFilm
//
//	@Summary		Обновление данных об фильме.
//	@Description	Обновляет данные об фильме. Все переданные поля будут обновлены. Отсутствующие поля /
//	будут оставлены без изменений.
//	@Tags			film
//	@Accept			json
//	@Param			film_id	path	uint64				true	"Уникальный идентификатор фильма"
//	@Param			request	body	request.UpdateFilm	true	"Информация об обновлении"
//	@Produce		json
//	@Success		200	{object}	response.Film		"Фильм успешно обновлен в базе"
//	@Failure		400	{object}	operate.ModelError	"В теле запроса ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на обновление фильма"
//	@Failure		404	{object}	operate.ModelError	"Фильм с указанным id не найден"
//	@Failure		409	{object}	operate.ModelError	"Актёр фильма не найден"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/film/{film_id} [put]
//	@Security		sessionCookie
func (fh *FilmHandlers) UpdateFilm(w http.ResponseWriter, r *http.Request, params mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение уникального идентификатора
	id, err := params.GetUint64(FilmIdField)
	if err != nil {
		operate.SendError(w, errors.Wrapf(err, "try get film id"), http.StatusBadRequest, l)
		return
	}

	// Получение значения тела запроса
	var updateFilm request.UpdateFilm
	if code, err := parseRequestBody(r.Body, &updateFilm, request.ValidateUpdateFilm, l); err != nil {
		operate.SendError(w, err, code, l)
		return
	}

	toUpdateFilm := &film.UpdateFilm{
		ID:           types.Id(id),
		Name:         updateFilm.Name,
		Description:  updateFilm.Description,
		DataPublish:  updateFilm.DataPublish,
		Rating:       updateFilm.Rating,
		UpdateActors: updateFilm.Actors != nil,
		Actors:       nil,
	}

	if updateFilm.Actors != nil {
		toUpdateFilm.Actors = *updateFilm.Actors
	}

	updatedFilm, err := fh.repository.UpdateFilm(toUpdateFilm)
	if err != nil {
		if errors.Is(err, film.ErrorFilmNotFound) {
			operate.SendError(w, ErrorFilmNotFound, http.StatusNotFound, l)
			return
		}

		if errors.Is(err, film.ErrorActorNotFound) {
			operate.SendError(w, ErrorActorNotFound, http.StatusConflict, l)
			l.Info(err)
			return
		}

		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't create film"))
		return
	}

	operate.SendStatus(w, http.StatusOK, response.FromRepositoryFilmWithActor(updatedFilm), l)
}
