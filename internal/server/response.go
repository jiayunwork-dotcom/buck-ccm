package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"buck-ccm/internal/spec"
)

// ErrorBody 是 API 错误响应的统一结构：
//
//	error  人类可读的错误文案（可被外行复述）
//	field  出错参数名（可选，校验类错误才有）
//
// 例如 {"error":"参数 d 非法：占空比必须在开区间 (0,1) 内（得到 1.5）","field":"d"}
type ErrorBody struct {
	Error string `json:"error"`
	Field string `json:"field,omitempty"`
}

// writeJSON 以指定状态码写 JSON 响应。
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 把任意 error 转成 ErrorBody 写出。
// 若 error 是 *spec.FieldError，额外带上 field 字段。
func writeError(w http.ResponseWriter, code int, err error) {
	err = flattenValidErr(err)
	body := ErrorBody{Error: err.Error()}
	var fe *spec.FieldError
	if errors.As(err, &fe) {
		body.Field = fe.Field
	}
	writeJSON(w, code, body)
}
