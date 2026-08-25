package server

import "net/http"

const (
	DefaultWebDir     = "web"
	DefaultExampleDir = "example"
)

func New() http.Handler {
	return NewWithDirs(DefaultWebDir, DefaultExampleDir)
}

func NewWithDirs(webDir, exampleDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/mode", handleMode)
	mux.HandleFunc("/api/ripple", handleRipple)
	mux.HandleFunc("/api/design", handleDesign)
	mux.HandleFunc("/api/sweep", handleSweep)
	mux.HandleFunc("/api/check", handleCheck)
	mux.HandleFunc("/api/boundary", handleBoundary)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/info", handleInfo)
	mux.HandleFunc("/api/endpoints", handleEndpoints)
	mux.Handle("/example/", http.StripPrefix("/example/",
		http.FileServer(http.Dir(exampleDir))))
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	return mux
}
