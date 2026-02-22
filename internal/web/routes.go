package web

import (
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/static"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.Files)))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		HealthPage().Render(r.Context(), w)
	})
	cop := http.NewCrossOriginProtection()
	return cop.Handler(mux)
}
