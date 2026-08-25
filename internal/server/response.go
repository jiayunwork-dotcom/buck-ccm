package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"buck-ccm/internal/spec"
)

type ErrorBody struct {
	Error string `json:"error"`
	Field string `json:"field,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, err error) {
	body := ErrorBody{Error: err.Error()}
	var fe *spec.FieldError
	if errors.As(err, &fe) {
		body.Field = fe.Field
	}
	writeJSON(w, code, body)
}
