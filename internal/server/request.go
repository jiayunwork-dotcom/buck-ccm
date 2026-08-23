package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"buck-ccm/internal/spec"
)

// maxBodyBytes 限制请求体大小，防止无界读入。
const maxBodyBytes = 1 << 20 // 1 MiB

// errHandled 表示错误响应已经写出，调用方只需 return。
var errHandled = errors.New("response already written")

// decodeSpec 解析 API 请求体。
//
// 失败路径都直接写 error JSON 并返回 errHandled：
//
//	非 POST 方法          → 405
//	请求体过大/读失败      → 400
//	空体/JSON 非法/未知字段 → 400
//	参数非法               → 400（校验文案来自 spec）
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
		writeError(w, http.StatusBadRequest, err)
		return nil, errHandled
	}
	return s, nil
}
