package app

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"io"
	"log"
	"net/http"
	"os"
	"vk_film/config"
	v1 "vk_film/internal/delivery/http/v1"
	"vk_film/internal/delivery/http/v1/handlers"
	"vk_film/internal/delivery/middleware"
	"vk_film/internal/pkg/prepare"
	"vk_film/internal/usecase/auth"
	"vk_film/pkg/logger"
	"vk_film/pkg/mux"

	_ "vk_film/docs"
)

func prepareLogger(cfg config.LoggerInfo) (*logger.Logger, *os.File) {
	var logOut io.Writer
	var logFile *os.File
	var err error

	if cfg.Directory != "" {
		logFile, err = prepare.OpenLogDir(cfg.Directory)
		if err != nil {
			log.Fatalf("[App] Init - create logger error: %s", err)
		}

		logOut = logFile
	} else {
		logOut = os.Stderr
		logFile = nil
	}

	l := logger.New(
		logger.Params{
			AppName:                  cfg.AppName,
			LogDir:                   cfg.Directory,
			Level:                    cfg.Level,
			UseStdAndFile:            cfg.UseStdAndFile,
			AddLowPriorityLevelToCmd: cfg.AllowShowLowLevel,
		},
		logOut,
	)

	return l, logFile
}

func Swagger(w http.ResponseWriter, r *http.Request, _ mux.Params) {
	httpSwagger.Handler()(w, r)
}

func prepareRoutes(actorHandlers *handlers.ActorHandlers, userHandlers *handlers.UserHandlers,
	filmHandlers *handlers.FilmHandlers, sessionManager auth.Manager) v1.Routes {
	return v1.Routes{
		//"Index"
		v1.Route{
			Method:      http.MethodGet,
			Pattern:     "/swagger/",
			HandlerFunc: Swagger,
		},

		v1.Route{
			Method:      http.MethodPost,
			Pattern:     "/actor",
			HandlerFunc: middleware.CheckSession(sessionManager)(actorHandlers.CreateActor),
		},

		// "DeleteActor"
		v1.Route{
			Method:      http.MethodDelete,
			Pattern:     "/actor/{" + handlers.ActorIdField + "}",
			HandlerFunc: middleware.CheckSession(sessionManager)(actorHandlers.DeleteActor),
		},

		// "GetActors"
		v1.Route{
			Method:      http.MethodGet,
			Pattern:     "/actor/list",
			HandlerFunc: middleware.CheckSession(sessionManager)(actorHandlers.GetActors),
		},

		// "UpdateActor"
		v1.Route{
			Method:      http.MethodPut,
			Pattern:     "/actor/{" + handlers.ActorIdField + "}",
			HandlerFunc: middleware.CheckSession(sessionManager)(actorHandlers.UpdateActor),
		},

		// "CreateFilm"
		v1.Route{
			Method:      http.MethodPost,
			Pattern:     "/film",
			HandlerFunc: middleware.CheckSession(sessionManager)(filmHandlers.CreateFilm),
		},

		// "DeleteFilm"
		v1.Route{
			Method:      http.MethodDelete,
			Pattern:     "/film/{" + handlers.FilmIdField + "}",
			HandlerFunc: middleware.CheckSession(sessionManager)(filmHandlers.DeleteFilm),
		},

		// "GetFilms",
		v1.Route{
			Method:      http.MethodGet,
			Pattern:     "/film/list",
			HandlerFunc: middleware.CheckSession(sessionManager)(filmHandlers.GetFilms),
		},

		// "UpdateFilm"
		v1.Route{
			Method:      http.MethodPut,
			Pattern:     "/film/{" + handlers.FilmIdField + "}",
			HandlerFunc: middleware.CheckSession(sessionManager)(filmHandlers.UpdateFilm),
		},

		// "CreateUser"
		v1.Route{
			Method:      http.MethodPost,
			Pattern:     "/user",
			HandlerFunc: middleware.CheckSession(sessionManager)(userHandlers.CreateUser),
		},

		// "DeleteUser"
		v1.Route{
			Method:      http.MethodDelete,
			Pattern:     "/user/{" + handlers.UserIdField + "}",
			HandlerFunc: middleware.CheckSession(sessionManager)(userHandlers.DeleteUser),
		},

		// "Login"
		v1.Route{
			Method:      http.MethodPost,
			Pattern:     "/login",
			HandlerFunc: middleware.CheckNoSession(sessionManager)(userHandlers.Login),
		},

		// "Logout"
		v1.Route{
			Method:      http.MethodPost,
			Pattern:     "/logout",
			HandlerFunc: middleware.CheckSession(sessionManager)(userHandlers.Logout),
		},

		// "UpdateUserRole"
		v1.Route{
			Method:      http.MethodPut,
			Pattern:     "/user/{" + handlers.UserIdField + "}/role",
			HandlerFunc: middleware.CheckSession(sessionManager)(userHandlers.UpdateUserRole),
		},

		// "GetUsers"
		v1.Route{
			Method:      http.MethodGet,
			Pattern:     "/user/list",
			HandlerFunc: middleware.CheckSession(sessionManager)(userHandlers.GetUsers),
		},
	}
}
