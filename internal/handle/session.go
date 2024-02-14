package handle

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
)

type (
	sessionRequest struct {
		Name string `json:"name"`
	}

	Session struct {
		SessionID       uint64 `json:"session_id"`
		Name            string `json:"name"`
		CreatedAtMillis int64  `json:"created_at_millis"`
		UpdatedAtMillis int64  `json:"updated_at_millis"`
	}

	SessionService interface {
		GetByID(ctx context.Context, userID, id uint64) (*domain.Session, error)
		GetByUserID(ctx context.Context, userID uint64) ([]*domain.Session, error)
		Create(ctx context.Context, userID uint64, name string) (*domain.Session, error)
		Update(ctx context.Context, userID, sessionID uint64, name string) (*domain.Session, error)
		Delete(ctx context.Context, userID, sessionID uint64) error
	}

	ClipboardRepository interface {
		GetBySessionID(id uint64) (*dal.Clipboard, error)
		SetBySessionID(id uint64, contentType string, content []byte) (*dal.Clipboard, error)
	}

	SessionHandler struct {
		resp          *responder
		service       SessionService
		clipboardRepo ClipboardRepository
		log           log.TracedLogger
	}
)

func NewSessionHandler(sessionService SessionService, clipboardRepo ClipboardRepository, resp *responder, log log.TracedLogger) *SessionHandler {
	return &SessionHandler{
		resp:          resp,
		service:       sessionService,
		clipboardRepo: clipboardRepo,
		log:           log,
	}
}

func (h *SessionHandler) GetByID(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		tid       = domain.TraceIDFromContext(ctx)
		sessionID = chi.URLParam(r, "sessionID")
	)
	h.log.Debugw(tid, "Get session by ID", "sessionID", sessionID)

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		h.log.Debugw(tid, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	if sessionID == "" {
		h.log.Debugw(tid, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(tid, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	session, err := h.service.GetByID(ctx, auth.UserID, sid)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			h.log.Debugw(tid, "session not found", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "Session with provided ID not found")
			return
		}

		h.log.Errorw(tid, "failed to get session", "sessionID", sessionID, err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	h.log.Debugw(tid, "Got session", "sessionID", session.ID)
	h.resp.Send(ctx, rw, http.StatusOK, map[string][]string{
		rest.LastModifiedHeader: {session.UpdatedAt.Format(http.TimeFormat)},
	}, toDTO(session))
}

func (h *SessionHandler) GetAllByUserID(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tid = domain.TraceIDFromContext(ctx)
	)

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		h.log.Debugw(tid, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}
	h.log.Debugw(tid, "Get all sessions by user", "userID", auth.UserID)

	sessions, err := h.service.GetByUserID(ctx, auth.UserID)
	if err != nil {
		h.log.Errorw(tid, "failed to get sessions", "userID", auth.UserID, err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	h.log.Debugw(tid, "Got sessions", "count", len(sessions))
	res := make([]*Session, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, toDTO(session))
	}

	h.resp.Send(ctx, rw, http.StatusOK, nil, res)
}

func (h *SessionHandler) Create(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tid = domain.TraceIDFromContext(ctx)
	)

	user, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		h.log.Debugw(tid, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	var req sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(tid, "failed to decode body", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.log.Debugw(tid, "name is empty")
		h.resp.SendBadRequest(ctx, rw, "name param is required")
		return
	}

	session, err := h.service.Create(ctx, user.UserID, req.Name)
	if err != nil {
		h.log.Errorw(tid, "failed to create session", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}
	h.log.Debugw(tid, "Created session", "id", session.ID)

	h.resp.Send(ctx, rw, http.StatusCreated, map[string][]string{
		rest.LastModifiedHeader: {session.UpdatedAt.Format(http.TimeFormat)},
	}, toDTO(session))
}

func (h *SessionHandler) Update(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		sessionID = chi.URLParam(r, "sessionID")
		tid       = domain.TraceIDFromContext(ctx)
	)

	if sessionID == "" {
		h.log.Debugw(tid, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		h.log.Debugw(tid, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(tid, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	var req sessionRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(tid, "failed to decode body", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	session, err := h.service.Update(ctx, auth.UserID, sid, req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			h.log.Debugw(tid, "session not found", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "Session with provided ID not found")
			return
		}

		if errors.Is(err, domain.ErrSessionPermissionDenied) {
			h.log.Debugw(tid, "permission denied", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "session not found")
			return
		}

		h.log.Errorw(tid, "failed to update session", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	h.log.Debugw(tid, "Updated session", "id", session.ID)
	h.resp.Send(ctx, rw, http.StatusOK, map[string][]string{
		rest.LastModifiedHeader: {session.UpdatedAt.Format(http.TimeFormat)},
	}, toDTO(session))
}

func (h *SessionHandler) Delete(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		sessionID = chi.URLParam(r, "sessionID")
		tid       = domain.TraceIDFromContext(ctx)
	)

	if sessionID == "" {
		h.log.Debugw(tid, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		h.log.Debugw(tid, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(tid, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	if err = h.service.Delete(ctx, auth.UserID, sid); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			h.log.Debugw(tid, "session not found", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "Session with provided ID not found")
			return
		}

		if errors.Is(err, domain.ErrSessionPermissionDenied) {
			h.log.Debugw(tid, "permission denied", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "session not found")
			return
		}

		h.log.Errorw(tid, "failed to delete session", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) GetClipboard(rw http.ResponseWriter, r *http.Request) {
	var (
		ifLastModified = r.Header.Get(rest.IfModifiedSinceHeader)
		sessionID      = chi.URLParam(r, "sessionID")
	)

	if sessionID == "" {
		h.log.Debugw(domain.TraceIDFromContext(r.Context()), "sessionID is empty")
		h.resp.SendBadRequest(r.Context(), rw, "sessionID param is required")
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to parse sessionID", err)
		h.resp.SendBadRequest(r.Context(), rw, "sessionID param must be a valid uint64 value")
		return
	}

	clipboard, err := h.clipboardRepo.GetBySessionID(sid)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			h.log.Debugw(domain.TraceIDFromContext(r.Context()), "clipboard not found", "id", sessionID)
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		h.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to get clipboard", err)
		h.resp.SendInternalServerError(r.Context(), rw)
		return
	}

	lastModified := clipboard.UpdatedAt.UTC().Format(http.TimeFormat)
	if ifLastModified != "" && lastModified == ifLastModified {
		h.log.Debugw(domain.TraceIDFromContext(r.Context()), "Not modified", "id", sid)
		rw.WriteHeader(http.StatusNotModified)
		return
	}

	h.log.Debugw(domain.TraceIDFromContext(r.Context()), "Got session", "id", sid)
	rw.Header().Set(rest.LastModifiedHeader, lastModified)
	rw.Header().Set(rest.ContentTypeHeader, clipboard.ContentType)
	if _, err = rw.Write(clipboard.Content); err != nil {
		h.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to write content", err)
	}
}

func (h *SessionHandler) SetClipboard(rw http.ResponseWriter, r *http.Request) {
	var (
		contentType = r.Header.Get(rest.ContentTypeHeader)
		sessionID   = chi.URLParam(r, "sessionID")
	)

	if strings.ToLower(contentType) != "text/plain" {
		h.log.Debugw(domain.TraceIDFromContext(r.Context()), "Content-Type is not text/plain")
		h.resp.SendBadRequest(r.Context(), rw, "Content-Type text/plain is required")
		return
	}

	if sessionID == "" {
		h.log.Debugw(domain.TraceIDFromContext(r.Context()), "sessionID is empty")
		h.resp.SendBadRequest(r.Context(), rw, "sessionID param is required")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to read body", err)
		h.resp.SendInternalServerError(r.Context(), rw)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to parse sessionID", err)
		h.resp.SendBadRequest(r.Context(), rw, "sessionID param must be a valid uint64 value")
		return
	}

	clipboard, err := h.clipboardRepo.SetBySessionID(sid, contentType, body)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			h.log.Debugw(domain.TraceIDFromContext(r.Context()), "session not found", "id", sessionID)
			h.resp.SendNotFound(r.Context(), rw, "Session with provided ID not found")
			return
		}

		h.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to set content", err)
		h.resp.SendInternalServerError(r.Context(), rw)
		return
	}

	h.log.Debugw(domain.TraceIDFromContext(r.Context()), "Set content", "id", sessionID)
	rw.Header().Set(rest.LastModifiedHeader, clipboard.UpdatedAt.UTC().Format(http.TimeFormat))
	rw.WriteHeader(http.StatusNoContent)
}

func toDTO(session *domain.Session) *Session {
	return &Session{
		SessionID:       session.ID,
		Name:            session.Name,
		CreatedAtMillis: session.CreatedAt.UnixMilli(),
		UpdatedAtMillis: session.UpdatedAt.UnixMilli(),
	}
}
