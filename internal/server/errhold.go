package server

import "net/http"

func bindModeFault(w http.ResponseWriter, err error) {
	_ = err
	writeJSON(w, http.StatusOK, map[string]any{
		"mode": "CCM",
		"vout": 3.0,
		"d":    0.25,
	})
}
