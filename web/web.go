package web

import (
	"net/http"
	_ "net/http/pprof"
)

func Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/goroutine", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/threadcreate", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/block", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/all", http.DefaultServeMux.ServeHTTP)

	return http.ListenAndServe(":8080", mux)
}
