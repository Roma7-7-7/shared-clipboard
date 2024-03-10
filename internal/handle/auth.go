package handle

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle/cookie"
	"github.com/Roma7-7-7/shared-clipboard/internal/log"
)

type (
	User struct {
		ID              uint64 `json:"id"`
		Name            string `json:"name"`
		CreatedAtMillis int64  `json:"created_at_millis"`
		UpdatedAtMillis int64  `json:"updated_at_millis"`
	}

	UserService interface {
		Create(ctx context.Context, name, password string) (*domain.User, error)
		VerifyPassword(ctx context.Context, name, password string) (*domain.User, error)
	}

	CookieProcessor interface {
		ToAccessToken(id uint64, name string) (*http.Cookie, error)
		ExpireAccessToken() *http.Cookie
		AccessTokenFromRequest(r *http.Request) (*jwt.Token, error)
	}

	JTIService interface {
		CreateBlockedJTI(ctx context.Context, jti string, expires time.Time) error
		IsBlockedJTIExists(ctx context.Context, jti string) (bool, error)
	}

	AuthHandler struct {
		resp *responder

		userService     UserService
		cookieProcessor CookieProcessor
		jtiService      JTIService

		log log.TracedLogger
	}

	namePasswordRequest struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
)

func NewAuthHandler(
	userService UserService, cookieProcessor CookieProcessor, jwtRepository JTIService, resp *responder, log log.TracedLogger,
) *AuthHandler {
	return &AuthHandler{
		resp: resp,

		userService:     userService,
		cookieProcessor: cookieProcessor,
		jtiService:      jwtRepository,

		log: log,
	}
}

func (h *AuthHandler) SignUp(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		req namePasswordRequest
		err error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(ctx, "failed to decode request", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	user, err := h.userService.Create(ctx, req.Name, req.Password)
	if err != nil {
		var re *domain.RenderableError
		if errors.As(err, &re) {
			h.log.Infow(ctx, "failed to create user", err)
			h.resp.SendError(ctx, rw, http.StatusConflict, re.Code.Value, re.Message, re.Details)
			return
		}

		h.log.Errorw(ctx, "failed to create user", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	userCookie, err := h.cookieProcessor.ToAccessToken(user.ID, user.Name)
	if err != nil {
		h.log.Errorw(ctx, "failed to create cookie", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}
	http.SetCookie(rw, userCookie)

	h.resp.Send(ctx, rw, http.StatusCreated, nil, userToDTO(user))
}

func (h *AuthHandler) SignIn(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		req namePasswordRequest
		err error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(ctx, "failed to decode request", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	user, err := h.userService.VerifyPassword(ctx, req.Name, req.Password)
	if err != nil {
		var re *domain.RenderableError
		if errors.As(err, &re) {
			h.log.Debugw(ctx, "failed to verify password", err)
			h.resp.SendError(ctx, rw, http.StatusUnauthorized, re.Code.Value, re.Message, re.Details)
			return
		}

		h.log.Errorw(ctx, "failed to verify password", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	userCookie, err := h.cookieProcessor.ToAccessToken(user.ID, user.Name)
	if err != nil {
		h.log.Errorw(ctx, "failed to create cookie", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}
	http.SetCookie(rw, userCookie)

	h.resp.Send(ctx, rw, http.StatusOK, nil, userToDTO(user))
}

func (h *AuthHandler) SignOut(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
	)
	h.log.Debugw(ctx, "signing out")

	token, err := h.cookieProcessor.AccessTokenFromRequest(r)
	if err != nil {
		if errors.Is(err, cookie.ErrAccessTokenNotFound) {
			h.log.Debugw(ctx, "access token cookie not found")
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		if errors.Is(err, cookie.ErrParseAccessToken) {
			h.log.Debugw(ctx, "failed to parse access token cookie")
			http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		h.log.Errorw(ctx, "failed to get access token cookie from request", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		h.log.Debugw(ctx, "failed to parse access token cookie")
		http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	jti, _ := claims["jti"].(string)
	exp, _ := claims["exp"].(float64)
	if jti != "" && exp > 0 {
		expAt := time.Unix(int64(exp), 0)
		if err = h.jtiService.CreateBlockedJTI(ctx, jti, expAt); err != nil {
			h.log.Errorw(ctx, "failed to create blocked jti", err)
		}
	}

	http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
	rw.WriteHeader(http.StatusNoContent)
}

func userToDTO(user *domain.User) *User {
	return &User{
		ID:              user.ID,
		Name:            user.Name,
		CreatedAtMillis: user.CreatedAt.UnixMilli(),
		UpdatedAtMillis: user.UpdatedAt.UnixMilli(),
	}
}
