package handle

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle/cookie"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

type AuthorizedMiddleware struct {
	resp            *responder
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
	cookieProcessor CookieProcessor, jwtRepository JWTRepository, resp *responder, log log.TracedLogger,
) *AuthorizedMiddleware {
	return &AuthorizedMiddleware{
		resp:            resp,
		cookieProcessor: cookieProcessor,
		jwtRepository:   jwtRepository,
		log:             log,
	}
}

func (m *AuthorizedMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var (
			ctx   = r.Context()
			tid   = domain.TraceIDFromContext(ctx)
			token *jwt.Token
			err   error
		)
		m.log.Debugw(tid, "authorized middleware")

		if token, err = m.cookieProcessor.AccessTokenFromRequest(r); err != nil {
			if errors.Is(err, cookie.ErrAccessTokenNotFound) {
				m.log.Debugw(tid, "access token cookie not found")
				m.resp.SendError(ctx, rw, http.StatusUnauthorized, domain.ErrorCodeUnauthorized.Value, "Request is not authorized", nil)
				return
			}
			if errors.Is(err, cookie.ErrParseAccessToken) {
				m.log.Debugw(tid, "failed to parse access token cookie")
				m.sendForbidden(ctx, rw, "JWT token is not valid or expired")
				return
			}

			m.log.Errorw(tid, "failed to get access token cookie from request", err)
			m.resp.SendInternalServerError(ctx, rw)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			m.log.Debugw(tid, "failed to parse access token cookie")
			m.sendForbidden(ctx, rw, "JWT token is not valid or expired")
			return
		}

		authority, err := toAuthority(claims)
		if err != nil {
			m.log.Errorw(tid, "failed to parse authority", err)
			m.resp.SendInternalServerError(ctx, rw)
			return
		}

		jti, ok := claims["jti"].(string)
		if ok && jti != "" {
			ok, err = m.jwtRepository.IsBlockedJTIExists(jti)
			if err != nil {
				m.log.Errorw(tid, "failed to check blocked jti", err)
				m.resp.SendInternalServerError(ctx, rw)
				return
			}
			if ok {
				m.log.Debugw(tid, "blocked jti")
				m.sendForbidden(ctx, rw, "JWT token is not valid or expired")
				return
			}
		}

		next.ServeHTTP(rw, r.WithContext(domain.ContextWithAuthority(ctx, authority)))
	})
}

func (m *AuthorizedMiddleware) sendForbidden(ctx context.Context, rw http.ResponseWriter, message string) {
	m.resp.SendError(ctx, rw, http.StatusForbidden, domain.ErrorCodeForbidden.Value, message, nil)
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
