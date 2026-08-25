package server

import (
	"net/http"

	"buck-ccm/internal/engine"
)

func handleMode(w http.ResponseWriter, r *http.Request) {
	s, err := decodeSpec(w, r)
	if err != nil {
		return
	}
	res, err := engine.Analyze(*s)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, res.ModeJSON())
}
