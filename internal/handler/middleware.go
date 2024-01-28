package handler

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/cookie"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

type AuthorizedMiddleware struct {
	cookieProcessor CookieProcessor
	jwtRepository   JWTRepository
	log             log.TracedLogger
}

func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid := r.Header.Get(middleware.RequestIDHeader)
		if tid == "" {
			tid = randomAlphanumericTraceID()
		}
		w.Header().Set(middleware.RequestIDHeader, tid)
		next.ServeHTTP(w, r.WithContext(domain.ContextWithTraceID(r.Context(), tid)))
	})
}

func Logger(l log.TracedLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			started := time.Now()
			defer func() {
				l.Infow(domain.TraceIDFromContext(r.Context()), "request",
					"method", r.Method,
					"url", r.URL.String(),
					"proto", r.Proto,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration", time.Since(started))
			}()

			next.ServeHTTP(ww, r)
		})
	}

}

func NewAuthorizedMiddleware(
	cookieProcessor CookieProcessor, jwtRepository JWTRepository, log log.TracedLogger,
) *AuthorizedMiddleware {
	return &AuthorizedMiddleware{
		cookieProcessor: cookieProcessor,
		jwtRepository:   jwtRepository,
		log:             log,
	}
}

func (m *AuthorizedMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var (
			ctx       = r.Context()
			tid       = domain.TraceIDFromContext(ctx)
			token     *jwt.Token
			claims    jwt.MapClaims
			ok        bool
			authority *domain.Authority
			err       error
		)
		m.log.Debugw(tid, "authorized middleware")

		if token, err = m.cookieProcessor.AccessTokenFromRequest(r); err != nil {
			if errors.Is(err, cookie.ErrAccessTokenNotFound) {
				m.log.Debugw(tid, "access token cookie not found")
				sendUnauthorized(ctx, rw, m.log)
				return
			}
			if errors.Is(err, cookie.ErrParseAccessToken) {
				m.log.Debugw(tid, "failed to parse access token cookie")
				sendForbidden(ctx, rw, "JWT token is not valid or expired", m.log)
				return
			}

			m.log.Errorw(tid, "failed to get access token cookie from request", err)
			sendInternalServerError(ctx, rw, m.log)
			return
		}

		if claims, ok = token.Claims.(jwt.MapClaims); !ok || !token.Valid {
			m.log.Debugw(tid, "failed to parse access token cookie")
			sendForbidden(ctx, rw, "JWT token is not valid or expired", m.log)
			return
		}

		if authority, err = toAuthority(claims); err != nil {
			m.log.Errorw(tid, "failed to parse authority", err)
			sendInternalServerError(ctx, rw, m.log)
			return
		}

		jti, ok := claims["jti"].(string)
		if ok && jti != "" {
			ok, err = m.jwtRepository.IsBlockedJTIExists(jti)
			if err != nil {
				m.log.Errorw(tid, "failed to check blocked jti", err)
				sendInternalServerError(ctx, rw, m.log)
				return
			}
			if ok {
				m.log.Debugw(tid, "blocked jti")
				sendForbidden(ctx, rw, "JWT token is not valid or expired", m.log)
				return
			}
		}

		next.ServeHTTP(rw, r.WithContext(domain.ContextWithAuthority(ctx, authority)))
	})
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomAlphanumericTraceID() string {
	b := make([]rune, 40)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func toAuthority(claims jwt.MapClaims) (*domain.Authority, error) {
	var (
		ids  string
		id   uint64
		name string
		ok   bool
		err  error
	)
	if ids, err = claims.GetSubject(); err != nil {
		return nil, fmt.Errorf("get subject: %w", err)
	}
	if id, err = strconv.ParseUint(ids, 10, 64); err != nil {
		return nil, fmt.Errorf("parse subject: %w", err)
	}
	if name, ok = claims["username"].(string); !ok {
		return nil, errors.New("name is not a string")
	}

	return &domain.Authority{
		UserID:   id,
		UserName: name,
	}, nil
}
