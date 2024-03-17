package handlers

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
	"vk_film/internal/delivery/http/v1/model/request"
	"vk_film/internal/delivery/http/v1/model/response"
	"vk_film/internal/delivery/middleware"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/user"
	"vk_film/internal/usecase/auth"
	"vk_film/pkg/mux"
	"vk_film/pkg/operate"
	"vk_film/pkg/slices"
)

const (
	DefaultRole = "user"
	UserIdField = "user_id"
)

type UserHandlers struct {
	repository user.Repository
	auth       auth.Manager
}

func NewUserHandlers(repository user.Repository, auth auth.Manager) *UserHandlers {
	return &UserHandlers{repository: repository, auth: auth}
}

// CreateUser
//
//	@Summary		Добавление пользователя.
//	@Description	Добавляет пользователя включая его логин, пароль и роль. По умолчанию роль 'user'.
//	@Tags			user
//	@Accept			json
//	@Param			request	body	request.CreateUser	true	"Информация о добавляемом пользователе"
//	@Produce		json
//	@Success		201	{object}	response.User		"Пользователь успешно добавлен в базу"
//	@Failure		400	{object}	operate.ModelError	"В теле запроса ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на создание пользователя"
//	@Failure		409	{object}	operate.ModelError	"Пользователь с таким же логином уже существует"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/user [post]
//	@Security		sessionCookie
func (uh *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение значения тела запроса
	var createUser request.CreateUser
	if code, err := parseRequestBody(r.Body, &createUser, request.ValidateCreateUser, l); err != nil {
		operate.SendError(w, err, code, l)
		return
	}

	// Шифруем пароль пользователя
	enc, err := bcrypt.GenerateFromPassword([]byte(createUser.Password), bcrypt.DefaultCost)
	if err != nil {
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Warn(errors.Wrapf(err, "try encrypt password for user"))
		return
	}

	if createUser.Role == "" {
		createUser.Role = DefaultRole
	}

	createdUser, err := uh.repository.CreateUser(&user.User{
		Login:    createUser.Login,
		Password: string(enc),
		Role:     types.Roles(createUser.Role),
	})
	if err != nil {
		if errors.Is(err, user.ErrorLoginAlreadyExists) {
			operate.SendError(w, ErrorUserAlreadyExists, http.StatusConflict, l)
			l.Info(errors.Wrapf(err, "can't create user"))
			return
		}
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't create user"))
		return
	}

	operate.SendStatus(w, http.StatusCreated, &response.User{
		ID:    createdUser.ID,
		Login: createdUser.Login,
		Role:  string(createdUser.Role),
	}, l)
}

// DeleteUser
//
//	@Summary		Удаление пользователя.
//	@Description	Удаляет пользователя по его id.
//	@Tags			user
//	@Param			user_id	path	uint64	true	"Уникальный идентификатор пользователя"
//	@Produce		json
//	@Success		200	"Пользователь успешно удалён"
//	@Failure		400	{object}	operate.ModelError	"В теле запроса ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на удаление пользователя"
//	@Failure		404	{object}	operate.ModelError	"Пользователь с указанным id не найден"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/user/{user_id} [delete]
//	@Security		sessionCookie
func (uh *UserHandlers) DeleteUser(w http.ResponseWriter, r *http.Request, params mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение уникального идентификатора
	id, err := params.GetUint64(UserIdField)
	if err != nil {
		operate.SendError(w, errors.Wrapf(err, "try get user id"), http.StatusBadRequest, l)
		return
	}

	if err = uh.repository.DeleteUser(types.Id(id)); err != nil {
		if errors.Is(err, user.ErrorUserNotFound) {
			operate.SendError(w, ErrorUserNotFound, http.StatusNotFound, l)
			return
		}
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't delete user"))
		return
	}

	operate.SendStatus(w, http.StatusOK, nil, l)
}

// UpdateUserRole
//
//	@Summary		Обновление роли пользователя.
//	@Description	Обновляет пользовательскую роль.
//	@Tags			user
//	@Accept			json
//	@Param			user_id	path	uint64				true	"Уникальный идентификатор пользователя"
//	@Param			request	body	request.UpdateRole	true	"Информация о добавляемом пользователе"
//	@Produce		json
//	@Success		200	{object}	response.User		"Роль пользователя успешно обновлена"
//	@Failure		400	{object}	operate.ModelError	"В теле запроса ошибка"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		403	{object}	operate.ModelError	"У пользователя нет прав на создание пользователя"
//	@Failure		404	{object}	operate.ModelError	"Пользователь с указанным id не найден"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/user/{user_id}/role [put]
//	@Security		sessionCookie
func (uh *UserHandlers) UpdateUserRole(w http.ResponseWriter, r *http.Request, params mux.Params) {
	l := middleware.GetLogger(r)

	// Проверка доступа
	if usr := middleware.GetUser(r); usr == nil || usr.Role != types.ADMIN {
		operate.SendError(w, ErrorUserNotPermitted, http.StatusForbidden, l)
		return
	}

	// Получение уникального идентификатора
	id, err := params.GetUint64(UserIdField)
	if err != nil {
		operate.SendError(w, errors.Wrapf(err, "try get user id"), http.StatusBadRequest, l)
		return
	}

	// Получение значения тела запроса
	var updateRole request.UpdateRole
	if code, err := parseRequestBody(r.Body, &updateRole, request.ValidateUpdateRole, l); err != nil {
		operate.SendError(w, err, code, l)
		return
	}

	updatedUser, err := uh.repository.UpdateUserRole(&user.User{
		ID:   types.Id(id),
		Role: types.Roles(updateRole.Role),
	})

	if err != nil {
		if errors.Is(err, user.ErrorUserNotFound) {
			operate.SendError(w, ErrorUserNotFound, http.StatusNotFound, l)
			return
		}
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't update user"))
		return
	}

	operate.SendStatus(w, http.StatusOK, &response.User{
		ID:    updatedUser.ID,
		Login: updatedUser.Login,
		Role:  string(updatedUser.Role),
	}, l)
}

// GetUsers
//
//	@Summary		Получение списка пользователей.
//	@Description	Возвращает список пользователей системы.
//	@Tags			user
//	@Produce		json
//	@Success		200	{array}		response.User		"Список пользователей успешно сформирован"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/user/list [get]
//	@Security		sessionCookie
func (uh *UserHandlers) GetUsers(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	users, err := uh.repository.GetUsers()
	if err != nil {
		operate.SendError(w, ErrorUnknownError, http.StatusInternalServerError, l)
		l.Error(errors.Wrapf(err, "can't get users"))
		return
	}

	operate.SendStatus(w, http.StatusOK, slices.Map(users, func(usr user.User) response.User {
		return response.User{
			ID:    usr.ID,
			Login: usr.Login,
			Role:  string(usr.Role),
		}
	}), l)
}

// Login
//
//	@Summary		Авторизация.
//	@Description	Авторизация пользователя в системе.
//	@Tags			user
//	@Accept			json
//	@Param			request	body	request.Login	true	"Логин и пароль пользователя"
//	@Produce		json
//	@Success		200	"Пользователь успешно авторизован"
//	@Header			200	{string}	Set-Cookie			"Устанавливает сессию текущего пользователя"
//	@Failure		400	{object}	operate.ModelError	"В теле запроса ошибка"
//	@Failure		409	{object}	operate.ModelError	"Неверный логин или пароль"
//	@Failure		418	{object}	operate.ModelError	"Пользователь уже авторизован"
//	@Failure		500	{object}	operate.ModelError	"Ошибка сервера"
//	@Router			/login [post]
func (uh *UserHandlers) Login(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	// Получение значения тела запроса
	var login request.Login
	if code, err := parseRequestBody(r.Body, &login, request.ValidateLogin, l); err != nil {
		operate.SendError(w, err, code, l)
		return
	}

	// Проверка верности логина и пароля
	sessionId, err := uh.auth.Login(login.Login, login.Password)
	if err != nil {
		if errors.Is(err, auth.ErrorIncorrectPassword) || errors.Is(err, user.ErrorUserNotFound) {
			operate.SendError(w, ErrorIncorrectLoginOrPassword, http.StatusConflict, l)
			l.Info(errors.Wrapf(err, "inccorect login info"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			l.Error(errors.Wrapf(err, "can't check login info"))
		}
		return
	}

	cookie := &http.Cookie{
		Name:     string(middleware.SessionField),
		Value:    sessionId,
		Path:     "/",
		Expires:  time.Now().Add(auth.ExpiredSessionTime),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	operate.SendStatus(w, http.StatusOK, nil, l)
}

// Logout
//
//	@Summary		Выход из системы.
//	@Description	Позволяет выйти пользователю из системы.
//	@Tags			user
//	@Produce		json
//	@Success		200	"Пользователь успешно вышел из системы"
//	@Failure		401	{object}	operate.ModelError	"Пользователь не авторизован"
//	@Router			/logout [post]
//	@Security		sessionCookie
func (uh *UserHandlers) Logout(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	l := middleware.GetLogger(r)

	sessionId := middleware.GetSession(r)
	if sessionId != nil {
		// Уничтожение сессии
		if err := uh.auth.Logout(*sessionId); err != nil {
			l.Warn(errors.Wrapf(err, "try delete session id %s", *sessionId))
		} else {
			l.Info("logged out")
		}

		cookie := &http.Cookie{
			Name:     string(middleware.SessionField),
			Value:    "",
			Path:     "/",
			Expires:  time.Now().AddDate(0, 0, -1),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		operate.SendStatus(w, http.StatusOK, nil, l)

	}

	l.Panic("no session was found in logout")
}
