package handlers

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"io"
	"net/http"
	"strings"
	"testing"
	"vk_film/internal/delivery/http/v1/model/request"
)

const (
	invalidJson      = "{ asdas"
	unknownFieldJson = "{ \"field\" : 1 }"
)

func TestUtils(t *testing.T) {
	runner.Run(t, "testing parseRequestBody", func(t provider.T) {
		t.NewStep("Init test data")
		body := "{ \"login\": \"login\", \"password\": \"password\" }"
		incorrectBody := "{ \"login\": \"login\", \"password\": 1 }"

		t.WithNewStep("Correct execute", func(t provider.StepCtx) {
			t.NewStep("Init body")
			b := io.NopCloser(strings.NewReader(body))

			t.NewStep("Check result")
			var login request.Login
			code, err := parseRequestBody(b, &login, request.ValidateLogin, &emptyLogger{})

			t.Require().NoError(err)
			t.Require().Equal(http.StatusOK, code)
		})

		t.WithNewStep("Body error execute", func(t provider.StepCtx) {
			t.NewStep("Check result")
			var login request.Login
			code, err := parseRequestBody(errReader(1), &login, request.ValidateLogin, &emptyLogger{})

			t.Require().ErrorIs(err, ErrorCannotReadBody)
			t.Require().Equal(http.StatusInternalServerError, code)
		})

		t.WithNewStep("Invalid json execute", func(t provider.StepCtx) {
			t.NewStep("Init body")
			b := io.NopCloser(strings.NewReader(invalidJson))

			t.NewStep("Check result")
			var login request.Login
			code, err := parseRequestBody(b, &login, request.ValidateLogin, &emptyLogger{})

			t.Require().ErrorIs(err, ErrorIncorrectBodyContent)
			t.Require().Equal(http.StatusBadRequest, code)
		})

		t.WithNewStep("Invalid fields execute", func(t provider.StepCtx) {
			t.NewStep("Init body")
			b := io.NopCloser(strings.NewReader(unknownFieldJson))

			t.NewStep("Check result")
			var login request.Login
			code, err := parseRequestBody(b, &login, request.ValidateLogin, &emptyLogger{})

			t.Require().Error(err)
			t.Require().Equal(http.StatusBadRequest, code)
		})

		t.WithNewStep("Unmarshal error execute", func(t provider.StepCtx) {
			t.NewStep("Init body")
			b := io.NopCloser(strings.NewReader(incorrectBody))

			t.NewStep("Check result")
			var login request.Login
			code, err := parseRequestBody(b, &login, func([]byte) error { return nil }, &emptyLogger{})

			t.Require().ErrorIs(err, ErrorIncorrectBodyContent)
			t.Require().Equal(http.StatusBadRequest, code)
		})
	})
}
