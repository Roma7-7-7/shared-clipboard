package handle

import (
	"net/http"

	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

type (
	UserInfo struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}

	UserHandler struct {
		resp *responder
		log  log.TracedLogger
	}
)

func NewUserHandler(resp *responder, log log.TracedLogger) *UserHandler {
	return &UserHandler{
		resp: resp,
		log:  log,
	}
}

func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.log.Debugw(domain.TraceIDFromContext(ctx), "get user info")

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		h.log.Errorw(domain.TraceIDFromContext(ctx), "authority not found in context")
		h.resp.SendInternalServerError(ctx, w)
		return
	}

	h.resp.Send(ctx, w, http.StatusOK, nil, UserInfo{
		ID:   auth.UserID,
		Name: auth.UserName,
	})
}
