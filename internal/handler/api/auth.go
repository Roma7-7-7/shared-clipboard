package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api/cookie"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
)

type (
	User struct {
		ID              uint64 `json:"id"`
		Name            string `json:"name"`
		CreatedAtMillis int64  `json:"created_at_millis"`
		UpdatedAtMillis int64  `json:"updated_at_millis"`
	}

	UserService interface {
		Create(ctx context.Context, name, password string) (*dal.User, error)
		VerifyPassword(ctx context.Context, name, password string) (*dal.User, error)
	}

	CookieProcessor interface {
		ToAccessToken(id uint64, name string) (*http.Cookie, error)
		ExpireAccessToken() *http.Cookie
		AccessTokenFromRequest(r *http.Request) (*jwt.Token, error)
	}

	JWTRepository interface {
		CreateBlockedJTI(jti string, expires time.Time) error
		IsBlockedJTIExists(jti string) (bool, error)
	}

	AuthHandler struct {
		userService     UserService
		cookieProcessor CookieProcessor
		jwtRepository   JWTRepository

		log log.TracedLogger
	}

	AuthorizedMiddleware struct {
		cookieProcessor CookieProcessor
		jwtRepository   JWTRepository
		log             log.TracedLogger
	}

	namePasswordRequest struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
)

func NewAuthHandler(
	userService UserService, cookieProcessor CookieProcessor, jwtRepository JWTRepository, log log.TracedLogger,
) *AuthHandler {
	return &AuthHandler{
		userService:     userService,
		cookieProcessor: cookieProcessor,
		jwtRepository:   jwtRepository,

		log: log,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/signup", h.SignUp)
	r.Post("/signin", h.SignIn)
	r.Post("/signout", h.SignOut)
}

func (h *AuthHandler) SignUp(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx        = r.Context()
		tid        = domain.TraceIDFromContext(ctx)
		req        namePasswordRequest
		user       *dal.User
		userCookie *http.Cookie
		marshaled  []byte
		err        error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(tid, "failed to decode request", err)
		sendBadRequest(ctx, rw, "failed to parse request", h.log)
		return
	}

	if user, err = h.userService.Create(ctx, req.Name, req.Password); err != nil {
		var re *domain.RenderableError
		if errors.As(err, &re) {
			h.log.Debugw(tid, "failed to create user", err)
			sendRenderableError(ctx, re, rw, h.log)
			return
		}

		h.log.Errorw(tid, "failed to create user", err)
		sendInternalServerError(ctx, rw, h.log)
		return
	}

	if userCookie, err = h.cookieProcessor.ToAccessToken(user.ID, user.Name); err != nil {
		h.log.Errorw(tid, "failed to create cookie", err)
		sendInternalServerError(ctx, rw, h.log)
		return
	}
	http.SetCookie(rw, userCookie)

	if marshaled, err = json.Marshal(userToDTO(user)); err != nil {
		h.log.Errorw(tid, "failed to marshal response", err)
		sendErrorMarshalBody(ctx, rw, h.log)
		return
	}

	rest.Send(ctx, rw, http.StatusCreated, rest.ContentTypeJSON, marshaled, h.log)
}

func (h *AuthHandler) SignIn(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx        = r.Context()
		tid        = domain.TraceIDFromContext(ctx)
		req        namePasswordRequest
		user       *dal.User
		userCookie *http.Cookie
		marshaled  []byte
		err        error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(tid, "failed to decode request", err)
		sendBadRequest(ctx, rw, "failed to parse request", h.log)
		return
	}

	if user, err = h.userService.VerifyPassword(ctx, req.Name, req.Password); err != nil {
		var re *domain.RenderableError
		if errors.As(err, &re) {
			h.log.Debugw(tid, "failed to verify password", err)
			sendRenderableError(ctx, re, rw, h.log)
			return
		}

		h.log.Errorw(tid, "failed to verify password", err)
		sendInternalServerError(ctx, rw, h.log)
		return
	}

	if userCookie, err = h.cookieProcessor.ToAccessToken(user.ID, user.Name); err != nil {
		h.log.Errorw(tid, "failed to create cookie", err)
		sendInternalServerError(ctx, rw, h.log)
		return
	}
	http.SetCookie(rw, userCookie)

	if marshaled, err = json.Marshal(userToDTO(user)); err != nil {
		h.log.Errorw(tid, "failed to marshal response", err)
		sendErrorMarshalBody(ctx, rw, h.log)
		return
	}

	rest.Send(ctx, rw, http.StatusOK, rest.ContentTypeJSON, marshaled, h.log)
}

func (h *AuthHandler) SignOut(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		tid    = domain.TraceIDFromContext(ctx)
		token  *jwt.Token
		claims jwt.MapClaims
		ok     bool
		err    error
	)
	h.log.Debugw(tid, "signing out")

	if token, err = h.cookieProcessor.AccessTokenFromRequest(r); err != nil {
		if errors.Is(err, cookie.ErrAccessTokenNotFound) {
			h.log.Debugw(tid, "access token cookie not found")
			rest.SendNoContent(ctx, rw, h.log)
			return
		}
		if errors.Is(err, cookie.ErrParseAccessToken) {
			h.log.Debugw(tid, "failed to parse access token cookie")
			http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
			rest.SendNoContent(ctx, rw, h.log)
			return
		}

		h.log.Errorw(tid, "failed to get access token cookie from request", err)
		sendInternalServerError(ctx, rw, h.log)
		return
	}

	if claims, ok = token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		h.log.Debugw(tid, "failed to parse access token cookie")
		http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
		rest.SendNoContent(ctx, rw, h.log)
		return
	}

	jti, _ := claims["jti"].(string)
	exp, _ := claims["exp"].(float64)
	if jti != "" && exp > 0 {
		expAt := time.Unix(int64(exp), 0)
		if err = h.jwtRepository.CreateBlockedJTI(jti, expAt); err != nil {
			h.log.Errorw(tid, "failed to create blocked jti", err)
		}
	}

	http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
	rest.SendNoContent(ctx, rw, h.log)
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

func userToDTO(user *dal.User) *User {
	return &User{
		ID:              user.ID,
		Name:            user.Name,
		CreatedAtMillis: user.CreatedAt.UnixMilli(),
		UpdatedAtMillis: user.UpdatedAt.UnixMilli(),
	}
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
