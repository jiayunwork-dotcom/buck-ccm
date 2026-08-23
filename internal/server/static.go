package server

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// EnsureStaticDirs 检查 web 与 example 目录存在，避免 HTTP 启动后静态
// 资源 404。返回可读错误，供 main 在启动前调用。
func EnsureStaticDirs(webDir, exampleDir string) error {
	if _, err := os.Stat(webDir); err != nil {
		return fmt.Errorf("web 目录 %s 不可访问：%v", webDir, err)
	}
	if _, err := os.Stat(exampleDir); err != nil {
		return fmt.Errorf("example 目录 %s 不可访问：%v", exampleDir, err)
	}
	return nil
}

// LoggingHandler 把每个请求的路径与耗时写到 stdout，便于本地观察
// API 调用（仅调试用，不改变业务行为）。
func LoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s (%s)\n", r.Method, r.URL.Path, time.Since(start))
	})
}
