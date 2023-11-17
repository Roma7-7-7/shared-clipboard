package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
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
	}

	CookieProcessor interface {
		ToAccessToken(id uint64, name string) (*http.Cookie, error)
	}

	AuthHandler struct {
		userService     UserService
		cookieProcessor CookieProcessor

		log log.TracedLogger
	}

	signUpRequest struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
)

func NewAuthHandler(userService UserService, cookieProcessor CookieProcessor, log log.TracedLogger) *AuthHandler {
	return &AuthHandler{
		userService:     userService,
		cookieProcessor: cookieProcessor,

		log: log,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/signup", h.SignUp)
}

func (h *AuthHandler) SignUp(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx        = r.Context()
		tid        = trace.ID(ctx)
		req        signUpRequest
		user       *dal.User
		userCookie *http.Cookie
		marshaled  []byte
		err        error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warnw(tid, "failed to decode request", err)
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

func userToDTO(user *dal.User) *User {
	return &User{
		ID:              user.ID,
		Name:            user.Name,
		CreatedAtMillis: user.CreatedAt.UnixMilli(),
		UpdatedAtMillis: user.UpdatedAt.UnixMilli(),
	}
}
