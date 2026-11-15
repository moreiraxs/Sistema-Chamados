package main

import (
	"log"
	"net/http"
	"time"
)

type ctxKey struct{}

type Identidade struct {
	ID   string
	Nome string
	EhTI bool
}

func identidadeDe(r *http.Request) Identidade {
	ident, _ := r.Context().Value(ctxKey{}).(Identidade)
	return ident
}

// restricao devolve o ID a filtrar: usuário comum só enxerga os próprios chamados; a TI enxerga todos.
func restricao(i Identidade) string {
	if i.EhTI {
		return ""
	}
	return i.ID
}

func cors(origem string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", origem)
		h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		h.Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func registrar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ini := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(ini).Round(time.Millisecond))
	})
}
