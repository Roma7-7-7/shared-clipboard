package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
)

type (
	UserInfo struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}

	UserHandler struct {
		log log.TracedLogger
	}
)

func NewUserHandler(log log.TracedLogger) *UserHandler {
	return &UserHandler{
		log: log,
	}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Get("/info", h.GetUserInfo)
}

func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.log.Debugw(domain.TraceIDFromContext(ctx), "get user info")

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		h.log.Errorw(domain.TraceIDFromContext(ctx), "authority not found in context")
		sendInternalServerError(ctx, w, h.log)
		return
	}

	body, err := rest.ToJSON(UserInfo{
		ID:   auth.UserID,
		Name: auth.UserName,
	})
	if err != nil {
		h.log.Errorw(domain.TraceIDFromContext(ctx), "marshal body", err)
		sendInternalServerError(ctx, w, h.log)
		return
	}

	rest.Send(ctx, w, http.StatusOK, rest.ContentTypeJSON, body, h.log)
}
