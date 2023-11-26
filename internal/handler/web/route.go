package web

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
)

func NewRouter(conf config.Web, log log.TracedLogger) (*chi.Mux, error) {
	log.Infow(domain.RuntimeTraceID, "Initializing web router")

	r := chi.NewRouter()

	if conf.Dev {
		r.Use(handler.Logger(log))
	}
	r.Use(httprate.LimitByIP(10, 1*time.Second))
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "text/html", "text/css", "text/javascript"))

	if conf.APIHost != "" {
		r.Get("/assets/js/env.js", envJson{
			lastModified: time.Now().Format(http.TimeFormat),
			response:     fmt.Sprintf("const apiHost = '%s';", conf.APIHost),
		}.Handle)
	}

	fs := http.Dir(conf.StaticFilesPath)
	r.Handle("/*", handleSpecificError{
		fs:       fs,
		delegate: http.FileServer(fs),
	})

	log.Infow(domain.RuntimeTraceID, "Router initialized")
	return r, nil
}

type envJson struct {
	lastModified string
	response     string
	log          log.TracedLogger
}

func (e envJson) Handle(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("If-Modified-Since") == e.lastModified {
		rw.WriteHeader(http.StatusNotModified)
		return
	}

	rw.Header().Set(rest.ContentTypeHeader, rest.ContentTypeJavaScript)
	rw.Header().Set(rest.LastModifiedHeader, e.lastModified)
	if _, err := rw.Write([]byte(e.response)); err != nil {
		e.log.Errorw(domain.TraceIDFromContext(r.Context()), "Failed to write response", err)
	}
}

type handleSpecificError struct {
	fs       http.FileSystem
	delegate http.Handler
}

func (h handleSpecificError) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	cleanPath := path.Clean(r.URL.Path)
	if cleanPath == "/404.html" { // to prevent infinite redirect loop
		h.delegate.ServeHTTP(rw, r)
		return
	}

	_, err := h.fs.Open(cleanPath)
	if os.IsNotExist(err) {
		http.Redirect(rw, r, "/404.html", http.StatusSeeOther)
		return
	}

	h.delegate.ServeHTTP(rw, r)
}
