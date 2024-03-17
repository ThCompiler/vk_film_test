package mux

import (
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"vk_film/pkg/logger"
)

type Params struct {
	req *http.Request
}

func NewParams(req *http.Request) *Params {
	return &Params{
		req: req,
	}
}

func (p *Params) GetUint64(name string) (uint64, error) {
	val, err := strconv.ParseUint(p.req.PathValue(name), 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "when try parse param \"%s\" int URL", name)
	}
	return val, nil
}

type ExtendedHandleFunc func(w http.ResponseWriter, r *http.Request, params Params)
type MiddlewareFunc func(ExtendedHandleFunc) ExtendedHandleFunc

type Mux struct {
	log         logger.Interface
	middlewares []MiddlewareFunc
	*http.ServeMux
}

func NewMux(log logger.Interface, middlewares ...MiddlewareFunc) *Mux {
	return &Mux{
		log:         log,
		middlewares: middlewares,
		ServeMux:    http.NewServeMux(),
	}
}

func (m *Mux) applyMiddleware(fun ExtendedHandleFunc) ExtendedHandleFunc {
	for _, middleware := range m.middlewares {
		fun = middleware(fun)
	}
	return fun
}

func (m *Mux) HandleFunc(method string, pattern string, fun ExtendedHandleFunc) {
	m.ServeMux.HandleFunc(method+" "+pattern, func(w http.ResponseWriter, r *http.Request) {
		m.applyMiddleware(fun)(w, r, Params{req: r})
	})
}
