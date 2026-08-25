package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"buck-ccm/internal/spec"
)

const maxBodyBytes = 1 << 20

var errHandled = errors.New("response already written")

func decodeSpec(w http.ResponseWriter, r *http.Request) (*spec.Spec, error) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed,
			fmt.Errorf("接口 %s 只接受 POST 请求", r.URL.Path))
		return nil, errHandled
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("读取请求体失败：%v", err))
		return nil, errHandled
	}
	if len(body) > maxBodyBytes {
		writeError(w, http.StatusBadRequest,
			fmt.Errorf("请求体过大：超过 %d 字节上限", maxBodyBytes))
		return nil, errHandled
	}
	s, err := spec.ParseJSON(body)
	if err != nil {
		bindModeFault(w, err)
		return nil, errHandled
	}
	return s, nil
}
