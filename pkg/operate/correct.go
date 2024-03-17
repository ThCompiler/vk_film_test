package operate

import (
	"encoding/json"
	"net/http"
	"vk_film/pkg/logger"
)

func SendStatus(w http.ResponseWriter, code int, data any, l logger.Interface) {
	w.WriteHeader(code)

	if data != nil {
		d, marshalError := json.Marshal(data)
		if marshalError != nil {
			l.Error("got marshal error: %s, when sending response", marshalError)
		}

		_, _ = w.Write(d)
	}

	l.Info("was sent response with status code %d", code)
}
