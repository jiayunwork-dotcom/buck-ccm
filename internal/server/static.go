package server

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func EnsureStaticDirs(webDir, exampleDir string) error {
	if _, err := os.Stat(webDir); err != nil {
		return fmt.Errorf("web 目录 %s 不可访问：%v", webDir, err)
	}
	if _, err := os.Stat(exampleDir); err != nil {
		return fmt.Errorf("example 目录 %s 不可访问：%v", exampleDir, err)
	}
	return nil
}

func LoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s (%s)\n", r.Method, r.URL.Path, time.Since(start))
	})
}
