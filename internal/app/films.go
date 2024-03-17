package app

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"os"
	"os/signal"
	"syscall"
	"vk_film/config"
	v1 "vk_film/internal/delivery/http/v1"
	"vk_film/internal/delivery/http/v1/handlers"
	"vk_film/internal/repository/actor"
	"vk_film/internal/repository/film"
	"vk_film/internal/repository/session"
	"vk_film/internal/repository/user"
	"vk_film/internal/usecase/auth"
	"vk_film/pkg/logger"
	"vk_film/pkg/server"

	_ "github.com/lib/pq"
)

func checkDatabaseConnections(db *sqlx.DB, rds *redis.Client, l logger.Interface) bool {
	if err := db.Ping(); err != nil {
		l.Fatal("[App] Init - can't check connection to sql with error %s", err)
		return false
	}

	l.Info("[App] Init - success check connection to postgresql")

	if err := rds.Ping(context.Background()).Err(); err != nil {
		l.Info("[App] Init - can't check connection to redis with error: %s", err)
		return false
	}

	l.Info("[App] Init - success check connection to redis")
	return true
}

func Run(cfg *config.Config) {
	// Logger
	l, logFile := prepareLogger(cfg.LoggerInfo)

	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
		_ = l.Sync()
	}()

	// Postgres
	pg, err := sqlx.Open("postgres", cfg.Postgres.URL)
	if err != nil {
		l.Fatal("[App] Init - postgres.New: %s", err)
	}
	defer pg.Close()

	// Redis
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		l.Fatal("[App] Init  - redis - redis.New: %s", err)
	}
	rds := redis.NewClient(opt)

	if !checkDatabaseConnections(pg, rds, l) {
		return
	}

	// Repository
	actorRepository := actor.NewPostgresActor(pg)
	userRepository := user.NewPostgresUser(pg)
	filmRepository := film.NewPostgresFilm(pg)
	sessionRepository := session.NewRedisSession(rds)

	// Use-cases
	sessionManager := auth.NewSessionManager(userRepository, sessionRepository)

	// Handlers
	actorHandlers := handlers.NewActorHandlers(actorRepository)
	userHandlers := handlers.NewUserHandlers(userRepository, sessionManager)
	filmHandlers := handlers.NewFilmHandlers(filmRepository)

	// routes
	router, err := v1.NewRouter("/api", l, prepareRoutes(actorHandlers, userHandlers, filmHandlers, sessionManager))
	if err != nil {
		l.Fatal("[App] Init - init handler error: %s", err)
	}

	httpServer := server.New(router, server.Port(cfg.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	l.Info("[App] Start - server started")

	select {
	case s := <-interrupt:
		l.Info("[App] Run - signal: " + s.String())
	case err := <-httpServer.Notify():
		l.Error(fmt.Errorf("[App] Run - httpServer.Notify: %s", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("[App] Stop - httpServer.Shutdown: %s", err))
	}

	l.Info("[App] Stop - server stopped")
}
