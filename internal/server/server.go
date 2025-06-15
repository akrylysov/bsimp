package server

import (
	"embed"
	"errors"
	"fmt"
	"github.com/ginqi7/bsimp/internal/auth"
	"github.com/ginqi7/bsimp/internal/config"
	"github.com/ginqi7/bsimp/internal/media"
	"html/template"
	"io/fs"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed templates static
var embedFS embed.FS

type Server struct {
	mediaLib      *media.MediaLibrary
	authLib       *auth.AuthLibrary
	tmpl          *template.Template
	staticVersion string
}

func httpError(r *http.Request, w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
	slog.Error("failed request",
		slog.String("error", err.Error()),
		slog.String("url", r.URL.String()),
		slog.Int("code", code),
	)
}

// ValidatePath provides a basic protection from the path traversal vulnerability.
func (s *Server) ValidatePath(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		error := s.authLib.CheckCookie(r)
		if error != nil {
			fmt.Println(error)
			http.Redirect(w, r, "/login_page/", http.StatusFound)
			// http.RedirectHandler("/login/", http.StatusMovedPermanently)
		}
		if strings.Contains(r.URL.Path, "./") || strings.Contains(r.URL.Path, ".\\") {
			httpError(r, w, errors.New("invalid path"), http.StatusBadRequest)
			return
		}
		h(w, r)
	}
}

// NormalizePath normalizes the request URL by removing the delimeter suffix.
func NormalizePath(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimRight(r.URL.Path, config.Delimiter)
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

type TemplateData struct {
	StaticVersion string
	*media.MediaListing
}

func (s *Server) ListingHandler(w http.ResponseWriter, r *http.Request) {
	listing, err := s.mediaLib.List(r.URL.Path)
	if err != nil {
		httpError(r, w, err, http.StatusInternalServerError)
		return
	}
	tmplData := TemplateData{
		StaticVersion: s.staticVersion,
		MediaListing:  listing,
	}
	if err := s.tmpl.ExecuteTemplate(w, "listing.gohtml", tmplData); err != nil {
		httpError(r, w, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmplData := TemplateData{
		StaticVersion: s.staticVersion,
	}
	if err := s.tmpl.ExecuteTemplate(w, "login.gohtml", tmplData); err != nil {
		httpError(r, w, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	password := r.FormValue("password")
	cookie := &http.Cookie{
		Name:    "password",
		Value:   password,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) StreamHandler(w http.ResponseWriter, r *http.Request) {
	url, err := s.mediaLib.ContentURL(r.URL.Path)
	if err != nil {
		httpError(r, w, err, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Server) AudioHandler(w http.ResponseWriter, r *http.Request) {
	// Open Audio File
	audioFile, err := s.mediaLib.OpenFile(r.URL.Path)
	defer audioFile.Close()
	if err != nil {
		http.Error(w, "Unable to open audio file", http.StatusInternalServerError)
		return
	}

	// Get File Info
	fileInfo, err := audioFile.Stat()
	if err != nil {
		http.Error(w, "Unable to get file info", http.StatusInternalServerError)
		return
	}
	rangeHeader := r.Header.Get("Range")

	if rangeHeader != "" {
		var start, end int64
		_, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
		if err != nil {
			_, err = fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
			end = 0
			if err != nil {
				http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
				return
			}
		}
		fileSize := fileInfo.Size()
		if start >= fileSize || end >= fileSize {
			http.Error(w, "Requested range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
			return
		}

		if end < 0 || end == 0 {
			end = fileSize - 1
		}

		chunkSize := end - start + 1

		// Set Content-Range
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		w.Header().Set("Content-Length", strconv.FormatInt(chunkSize, 10))
		w.WriteHeader(http.StatusPartialContent)
		audioFile.Seek(start, 0)
		buffer := make([]byte, chunkSize)
		audioFile.Read(buffer)
		w.Write(buffer)
	} else {
		// Set Content-Length
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		w.WriteHeader(http.StatusOK)
		http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), audioFile)
	}

	// Set Content-Type
	w.Header().Set("Content-Type", "audio/mpeg")
	// Set Accept-Ranges, for Chrome setting `currentTime'
	// reference: https://segmentfault.com/q/1010000002908474
	w.Header().Set("Accept-Ranges", "bytes")
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
func StartServer(mediaLib *media.MediaLibrary, authLib *auth.AuthLibrary, addr string) error {
	tmpl, err := template.New("").Funcs(templateFunctions).ParseFS(embedFS, "templates/*.gohtml")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	staticVersion := fmt.Sprintf("%x", rand.Uint64())
	staticFS, err := fs.Sub(embedFS, "static")
	if err != nil {
		return err
	}
	staticPath := fmt.Sprintf("/static/%s/", staticVersion)
	mux.Handle(staticPath, DisableFileListing(http.StripPrefix(staticPath, http.FileServer(http.FS(staticFS)))))

	s := Server{
		mediaLib:      mediaLib,
		authLib:       authLib,
		tmpl:          tmpl,
		staticVersion: staticVersion,
	}
	mux.Handle("/library/", http.StripPrefix("/library/", s.ValidatePath(NormalizePath(s.ListingHandler))))
	mux.Handle("/stream/", http.StripPrefix("/stream/", s.ValidatePath(NormalizePath(s.StreamHandler))))
	mux.Handle("/audio/", http.StripPrefix("/audio/", s.ValidatePath(NormalizePath(s.AudioHandler))))
	mux.Handle("/login_page/", http.StripPrefix("/login_page/", NormalizePath(s.LoginPage)))
	mux.Handle("/login", http.StripPrefix("/login", NormalizePath(s.LoginHandler)))
	mux.Handle("/", http.RedirectHandler("/library/", http.StatusMovedPermanently))

	return http.ListenAndServe(addr, mux)
}
