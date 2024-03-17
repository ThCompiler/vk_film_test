package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
	"time"
	"vk_film/internal/pkg/types"
	"vk_film/pkg/logger"
	"vk_film/pkg/mux"
)

const DataFormat = "2006/01/02 - 15:04:05"

const (
	RequestID logger.Field = "request_id"
	Method    logger.Field = "method"
	URL       logger.Field = "url"

	LoggerField types.ContextField = "logger"
)

func RequestLogger(l logger.Interface) mux.MiddlewareFunc {
	return func(fun mux.ExtendedHandleFunc) mux.ExtendedHandleFunc {
		return func(w http.ResponseWriter, r *http.Request, params mux.Params) {
			// Start timer
			start := time.Now()
			path := r.URL.Path
			raw := r.URL.RawQuery

			method := r.Method
			requestID := uuid.New()

			if raw != "" {
				path = path + "?" + raw
			}

			lg := l.With(URL, path).With(RequestID, requestID).With(Method, method)
			contextedRequest := r.WithContext(context.WithValue(r.Context(), LoggerField, lg))
			clientIP := r.RemoteAddr
			l.Info("[HTTP] Start - | %v | %s | %s  %v |",
				start.Format(DataFormat),
				clientIP,
				method,
				path,
			)

			// Process request
			fun(w, contextedRequest, params)

			// Stop timer
			timeStamp := time.Now()
			latency := timeStamp.Sub(start)

			if latency > time.Minute {
				latency = latency.Truncate(time.Second)
			}

			l.Info("[HTTP] End - | %v | %s | %s  %v | %v |",
				timeStamp.Format(DataFormat),
				clientIP,
				method,
				path,
				latency,
			)
		}
	}
}

func GetLogger(r *http.Request) logger.Interface {
	if lg := r.Context().Value(LoggerField); lg != nil {
		return lg.(logger.Interface)
	}

	return logger.DefaultLogger
}
