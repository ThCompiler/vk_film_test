package operate

import (
	"encoding/json"
	"net/http"
	"vk_film/pkg/logger"
)

type ModelError struct {
	ErrorMessage string `json:"error_message,omitempty"`
}

func SendError(w http.ResponseWriter, err error, code int, l logger.Interface) {
	w.WriteHeader(code)

	d, marshalError := json.Marshal(ModelError{ErrorMessage: err.Error()})
	if marshalError != nil {
		l.Error("got marshal error: %s, when send error: %s", marshalError, err)
	}

	_, _ = w.Write(d)
	l.Info("error %s was sent with status code %d", err, code)
}
