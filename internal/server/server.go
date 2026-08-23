// Package server 提供 buck-ccm 的薄 HTTP 层。
//
// 职责边界：把 POST /api/mode 与 POST /api/ripple 的 JSON 请求体解析成
// internal/spec 的算例，交给 internal/engine 核算，再把结果或错误体以
// JSON 返回；由同一进程挂载 web/（页面）与 example/（算例）静态资源。
// server 不包含任何求解逻辑，只做接线与错误表达。
package server

import "net/http"

// DefaultWebDir 与 DefaultExampleDir 是静态资源的相对目录（相对仓根）。
const (
	DefaultWebDir     = "web"
	DefaultExampleDir = "example"
)

// New 创建完整的 HTTP 处理器：
//
//	POST /api/mode     模式与稳态核算
//	POST /api/ripple   纹波核算与电感电流三角波点列
//	GET  /example/...  example 算例（供前端"加载示例"）
//	GET  /...          web/ 静态页面
func New() http.Handler {
	return NewWithDirs(DefaultWebDir, DefaultExampleDir)
}

// NewWithDirs 与 New 相同，但允许覆盖静态目录（用于测试与定制部署）。
func NewWithDirs(webDir, exampleDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/mode", handleMode)
	mux.HandleFunc("/api/ripple", handleRipple)
	mux.Handle("/example/", http.StripPrefix("/example/",
		http.FileServer(http.Dir(exampleDir))))
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	return mux
}
