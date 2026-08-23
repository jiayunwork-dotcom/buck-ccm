package server

import (
	"net/http"

	"buck-ccm/internal/engine"
)

// handleMode 实现 POST /api/mode。
//
// 请求体：{"vin":12,"d":0.4167,"l":1e-4,"c":2.2e-4,"ts":1e-5,"r":10}
// 成功返回 200 + 模式/稳态 JSON；参数非法返回 400 + error JSON。
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
