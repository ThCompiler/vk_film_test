package handlers

import (
	"context"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"vk_film/internal/delivery/middleware"
	"vk_film/internal/pkg/types"
	"vk_film/internal/repository/user"
	"vk_film/pkg/logger"
)

var testError = errors.New("test error")

var adminUser = &user.User{
	Role: types.ADMIN,
}

var userUser = &user.User{
	Role: types.USER,
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, testError
}

func (errReader) Close() error {
	return testError
}

type emptyLogger struct{}

func (*emptyLogger) Debug(_ any, _ ...any)                          {}
func (*emptyLogger) Info(_ any, _ ...any)                           {}
func (*emptyLogger) Warn(_ any, _ ...any)                           {}
func (*emptyLogger) Error(_ any, _ ...any)                          {}
func (*emptyLogger) Panic(_ any, _ ...any)                          {}
func (*emptyLogger) Fatal(_ any, _ ...any)                          {}
func (el *emptyLogger) With(_ logger.Field, _ any) logger.Interface { return el }

func initRequest(body io.Reader, contextValues map[types.ContextField]any) (*http.Request, error) {
	req, err := http.NewRequest("", "", body)

	if err != nil {
		return nil, err
	}

	ctx := req.Context()
	for k, v := range contextValues {
		ctx = context.WithValue(ctx, k, v)
	}
	ctx = context.WithValue(ctx, middleware.LoggerField, logger.Interface(&emptyLogger{}))

	return req.WithContext(ctx), nil
}
