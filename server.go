package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed templates static
var embedFS embed.FS

type Server struct {
	mediaLib *MediaLibrary
	tmpl     *template.Template
}

// ValidatePath provides a basic protection from the path traversal vulnerability.
func ValidatePath(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "./") || strings.Contains(r.URL.Path, ".\\") {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		h(w, r)
	}
}

// DisableFileListing disables file listing under directories. It can be used with the built-in http.FileServer.
func DisableFileListing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (s *Server) ListingHandler(w http.ResponseWriter, r *http.Request) {
	listing, err := s.mediaLib.List(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.tmpl.ExecuteTemplate(w, "listing.gohtml", listing); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) StreamHandler(w http.ResponseWriter, r *http.Request) {
	url, err := s.mediaLib.ContentURL(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

// Don't include sprig just for one function.
var templateFunctions = map[string]any{
	"defaultString": func(s string, def string) string {
		if s == "" {
			return def
		}
		return s
	},
}

// StartServer starts HTTP server.
func StartServer(mediaLib *MediaLibrary, addr string) error {
	tmpl, err := template.New("").Funcs(templateFunctions).ParseFS(embedFS, "templates/*.gohtml")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	staticFS, err := fs.Sub(embedFS, "static")
	if err != nil {
		return err
	}
	mux.Handle("/static/", DisableFileListing(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))))

	s := Server{
		mediaLib: mediaLib,
		tmpl:     tmpl,
	}
	mux.Handle("/library/", http.StripPrefix("/library/", ValidatePath(s.ListingHandler)))
	mux.Handle("/stream/", http.StripPrefix("/stream/", ValidatePath(s.StreamHandler)))

	return http.ListenAndServe(addr, mux)
}
