package server

import (
	"net/http"
)

type endpointInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Doc    string `json:"doc"`
}

type serviceInfo struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Routes  []string `json:"routes"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed,
			httpErrorf("接口 %s 只接受 GET 请求", r.URL.Path))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed,
			httpErrorf("接口 %s 只接受 GET 请求", r.URL.Path))
		return
	}
	writeJSON(w, http.StatusOK, serviceInfo{
		Name:    "buck-ccm",
		Version: "1.0.0",
		Routes: []string{
			"/api/mode", "/api/ripple", "/api/design",
			"/api/sweep", "/api/check", "/api/boundary",
			"/api/health", "/api/info", "/api/endpoints",
		},
	})
}

func handleEndpoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed,
			httpErrorf("接口 %s 只接受 GET 请求", r.URL.Path))
		return
	}
	writeJSON(w, http.StatusOK, []endpointInfo{
		{Method: "POST", Path: "/api/mode", Doc: "模式与稳态核算"},
		{Method: "POST", Path: "/api/ripple", Doc: "纹波与电感电流点列"},
		{Method: "POST", Path: "/api/design", Doc: "反求占空比并回读"},
		{Method: "POST", Path: "/api/sweep", Doc: "D 或 L 扫描"},
		{Method: "POST", Path: "/api/check", Doc: "交叉规则自查"},
		{Method: "POST", Path: "/api/boundary", Doc: "CCM/DCM 边界电感"},
		{Method: "GET", Path: "/api/health", Doc: "健康检查"},
		{Method: "GET", Path: "/api/info", Doc: "服务信息"},
		{Method: "GET", Path: "/api/endpoints", Doc: "路由列表"},
	})
}
