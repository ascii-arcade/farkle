package web

import (
	"html/template"
	"net/http"
	_ "net/http/pprof"

	"github.com/ascii-arcade/farkle/config"
	"github.com/ascii-arcade/farkle/games"
)

func Run() error {
	mux := http.NewServeMux()

	if config.GetDebug() {
		mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/goroutine", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/threadcreate", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/block", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
		mux.HandleFunc("/debug/pprof/all", http.DefaultServeMux.ServeHTTP)
	}

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/favicon.ico")
	})

	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web"+r.URL.Path)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		totalGames := len(games.GetAll())
		totalStartedGames := 0
		for _, game := range games.GetAll() {
			if game.Started {
				totalStartedGames++
			}
		}

		t, err := template.ParseFiles("web/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = t.Execute(w, struct {
			TotalGames        int
			TotalStartedGames int
		}{
			TotalGames:        totalGames,
			TotalStartedGames: totalStartedGames,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return http.ListenAndServe(":8080", mux)
}
