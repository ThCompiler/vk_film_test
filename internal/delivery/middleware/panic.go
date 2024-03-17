package middleware

import (
	"net/http"
	"runtime/debug"
	"vk_film/pkg/logger"
	"vk_film/pkg/mux"
)

func CheckPanic(fun mux.ExtendedHandleFunc) mux.ExtendedHandleFunc {
	return func(w http.ResponseWriter, r *http.Request, params mux.Params) {
		defer func(log logger.Interface, w http.ResponseWriter) {
			if err := recover(); err != nil {
				log.Error("detected critical error: %v, with stack: %s", err, debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)
			}
		}(GetLogger(r), w)

		// Process request
		fun(w, r, params)
	}
}
