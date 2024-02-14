package handle

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle/cookie"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
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
		resp *responder

		userService     UserService
		cookieProcessor CookieProcessor
		jwtRepository   JWTRepository

		log log.TracedLogger
	}

	namePasswordRequest struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
)

func NewAuthHandler(
	userService UserService, cookieProcessor CookieProcessor, jwtRepository JWTRepository, resp *responder, log log.TracedLogger,
) *AuthHandler {
	return &AuthHandler{
		resp: resp,

		userService:     userService,
		cookieProcessor: cookieProcessor,
		jwtRepository:   jwtRepository,

		log: log,
	}
}

func (h *AuthHandler) SignUp(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tid = domain.TraceIDFromContext(ctx)
		req namePasswordRequest
		err error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(tid, "failed to decode request", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	user, err := h.userService.Create(ctx, req.Name, req.Password)
	if err != nil {
		var re *domain.RenderableError
		if errors.As(err, &re) {
			h.log.Infow(tid, "failed to create user", err)
			h.resp.SendError(ctx, rw, http.StatusConflict, re.Code.Value, re.Message, re.Details)
			return
		}

		h.log.Errorw(tid, "failed to create user", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	userCookie, err := h.cookieProcessor.ToAccessToken(user.ID, user.Name)
	if err != nil {
		h.log.Errorw(tid, "failed to create cookie", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}
	http.SetCookie(rw, userCookie)

	h.resp.Send(ctx, rw, http.StatusCreated, nil, userToDTO(user))
}

func (h *AuthHandler) SignIn(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tid = domain.TraceIDFromContext(ctx)
		req namePasswordRequest
		err error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(tid, "failed to decode request", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	user, err := h.userService.VerifyPassword(ctx, req.Name, req.Password)
	if err != nil {
		var re *domain.RenderableError
		if errors.As(err, &re) {
			h.log.Debugw(tid, "failed to verify password", err)
			h.resp.SendError(ctx, rw, http.StatusUnauthorized, re.Code.Value, re.Message, re.Details)
			return
		}

		h.log.Errorw(tid, "failed to verify password", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	userCookie, err := h.cookieProcessor.ToAccessToken(user.ID, user.Name)
	if err != nil {
		h.log.Errorw(tid, "failed to create cookie", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}
	http.SetCookie(rw, userCookie)

	h.resp.Send(ctx, rw, http.StatusOK, nil, userToDTO(user))
}

func (h *AuthHandler) SignOut(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx   = r.Context()
		tid   = domain.TraceIDFromContext(ctx)
		token *jwt.Token
		err   error
	)
	h.log.Debugw(tid, "signing out")

	if token, err = h.cookieProcessor.AccessTokenFromRequest(r); err != nil {
		if errors.Is(err, cookie.ErrAccessTokenNotFound) {
			h.log.Debugw(tid, "access token cookie not found")
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		if errors.Is(err, cookie.ErrParseAccessToken) {
			h.log.Debugw(tid, "failed to parse access token cookie")
			http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		h.log.Errorw(tid, "failed to get access token cookie from request", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		h.log.Debugw(tid, "failed to parse access token cookie")
		http.SetCookie(rw, h.cookieProcessor.ExpireAccessToken())
		rw.WriteHeader(http.StatusNoContent)
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
	rw.WriteHeader(http.StatusNoContent)
}

func userToDTO(user *dal.User) *User {
	return &User{
		ID:              user.ID,
		Name:            user.Name,
		CreatedAtMillis: user.CreatedAt.UnixMilli(),
		UpdatedAtMillis: user.UpdatedAt.UnixMilli(),
	}
}
