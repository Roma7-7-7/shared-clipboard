package handler

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(middleware.RequestIDHeader)
		if tid == "" {
			tid = randomAlphanumericTraceID()
		}
		w.Header().Set(middleware.RequestIDHeader, tid)
		next.ServeHTTP(w, r.WithContext(trace.WithTraceID(r.Context(), tid)))
	})
}

func Logger(l trace.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			started := time.Now()
			defer func() {
				l.Infow(r.Context(), "request",
					"method", r.Method,
					"url", r.URL.String(),
					"proto", r.Proto,
					"remoteAddr", r.RemoteAddr,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration", time.Since(started))
			}()

			next.ServeHTTP(ww, r)
		})
	}

}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomAlphanumericTraceID() string {
	b := make([]rune, 40)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
