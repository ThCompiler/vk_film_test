package v1

import (
	"net/url"
	middleware2 "vk_film/internal/delivery/middleware"
	"vk_film/pkg/logger"
	"vk_film/pkg/mux"
)

const version = "v1"

type Route struct {
	Method      string
	Pattern     string
	HandlerFunc mux.ExtendedHandleFunc
}

type Routes []Route

func NewRouter(root string, l logger.Interface, routes Routes) (*mux.Mux, error) {
	router := mux.NewMux(l, middleware2.CheckPanic, middleware2.RequestLogger(l))

	for _, route := range routes {
		base, err := url.Parse(root)
		if err != nil {
			return nil, err
		}

		uRL := base.JoinPath(version, route.Pattern).Path

		router.HandleFunc(route.Method, uRL, route.HandlerFunc)
	}

	return router, nil
}
