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

	ac "github.com/Roma7-7-7/shared-clipboard/internal/context"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/log"
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
		FilterBy(ctx context.Context, userID uint64, filter domain.SessionFilter) ([]*domain.Session, int, error)
		Create(ctx context.Context, userID uint64, name string) (*domain.Session, error)
		Update(ctx context.Context, userID, sessionID uint64, name string) (*domain.Session, error)
		UpdateUpdatedAt(ctx context.Context, sessionID uint64) error
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
		sessionID = chi.URLParam(r, "sessionID")
	)
	h.log.Debugw(ctx, "Get session by ID", "sessionID", sessionID)

	auth, ok := ac.AuthorityFrom(ctx)
	if !ok {
		h.log.Debugw(ctx, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	if sessionID == "" {
		h.log.Debugw(ctx, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(ctx, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	session, err := h.service.GetByID(ctx, auth.UserID, sid)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			h.log.Debugw(ctx, "session not found", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "Session with provided ID not found")
			return
		}

		h.log.Errorw(ctx, "failed to get session", "sessionID", sessionID, err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	h.log.Debugw(ctx, "Got session", "sessionID", session.ID)
	h.resp.Send(ctx, rw, http.StatusOK, map[string][]string{
		LastModifiedHeader: {session.UpdatedAt.Format(http.TimeFormat)},
	}, toDTO(session))
}

func (h *SessionHandler) FilterBy(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
	)

	auth, ok := ac.AuthorityFrom(ctx)
	if !ok {
		h.log.Debugw(ctx, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}
	h.log.Debugw(ctx, "Get all sessions by user", "userID", auth.UserID)

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "100"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.log.Debugw(ctx, "failed to parse limit", "limit", limitStr, err)
		h.resp.SendBadRequest(ctx, rw, "limit param must be a valid int value")
		return
	}
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		h.log.Debugw(ctx, "failed to parse offset", "offset", offset, err)
		h.resp.SendBadRequest(ctx, rw, "offset param must be a valid int value")
		return
	}

	sessions, total, err := h.service.FilterBy(ctx, auth.UserID, domain.SessionFilter{
		Name:       r.URL.Query().Get("name"),
		SortBy:     r.URL.Query().Get("sortBy"),
		SortByDesc: strings.EqualFold(r.URL.Query().Get("desc"), "true"),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		h.log.Errorw(ctx, "failed to get sessions", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	h.log.Debugw(ctx, "Got sessions", "count", len(sessions))
	res := make([]*Session, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, toDTO(session))
	}

	h.resp.Send(ctx, rw, http.StatusOK, nil, &paginatedResponse{
		Items:      res,
		TotalItems: total,
	})
}

func (h *SessionHandler) Create(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
	)

	user, ok := ac.AuthorityFrom(ctx)
	if !ok {
		h.log.Debugw(ctx, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	var req sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(ctx, "failed to decode body", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.log.Debugw(ctx, "name is empty")
		h.resp.SendBadRequest(ctx, rw, "name param is required")
		return
	}

	session, err := h.service.Create(ctx, user.UserID, req.Name)
	if err != nil {
		h.log.Errorw(ctx, "failed to create session", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}
	h.log.Debugw(ctx, "Created session", "id", session.ID)

	h.resp.Send(ctx, rw, http.StatusCreated, map[string][]string{
		LastModifiedHeader: {session.UpdatedAt.Format(http.TimeFormat)},
	}, toDTO(session))
}

func (h *SessionHandler) Update(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		sessionID = chi.URLParam(r, "sessionID")
	)

	if sessionID == "" {
		h.log.Debugw(ctx, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	auth, ok := ac.AuthorityFrom(ctx)
	if !ok {
		h.log.Debugw(ctx, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(ctx, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	var req sessionRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debugw(ctx, "failed to decode body", err)
		h.resp.SendBadRequest(ctx, rw, "failed to parse request")
		return
	}

	session, err := h.service.Update(ctx, auth.UserID, sid, req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			h.log.Debugw(ctx, "session not found", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "Session with provided ID not found")
			return
		}

		if errors.Is(err, domain.ErrSessionPermissionDenied) {
			h.log.Debugw(ctx, "permission denied", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "session not found")
			return
		}

		h.log.Errorw(ctx, "failed to update session", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	h.log.Debugw(ctx, "Updated session", "id", session.ID)
	h.resp.Send(ctx, rw, http.StatusOK, map[string][]string{
		LastModifiedHeader: {session.UpdatedAt.Format(http.TimeFormat)},
	}, toDTO(session))
}

func (h *SessionHandler) Delete(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		sessionID = chi.URLParam(r, "sessionID")
	)

	if sessionID == "" {
		h.log.Debugw(ctx, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	auth, ok := ac.AuthorityFrom(ctx)
	if !ok {
		h.log.Debugw(ctx, "user not found in context")
		h.resp.SendUnauthorized(ctx, rw)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(ctx, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	if err = h.service.Delete(ctx, auth.UserID, sid); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			h.log.Debugw(ctx, "session not found", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "Session with provided ID not found")
			return
		}

		if errors.Is(err, domain.ErrSessionPermissionDenied) {
			h.log.Debugw(ctx, "permission denied", "sessionID", sessionID)
			h.resp.SendNotFound(ctx, rw, "session not found")
			return
		}

		h.log.Errorw(ctx, "failed to delete session", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) GetClipboard(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx            = r.Context()
		ifLastModified = r.Header.Get(IfModifiedSinceHeader)
		sessionID      = chi.URLParam(r, "sessionID")
	)

	if sessionID == "" {
		h.log.Debugw(ctx, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(ctx, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	clipboard, err := h.clipboardRepo.GetBySessionID(sid)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			h.log.Debugw(ctx, "clipboard not found", "id", sessionID)
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		h.log.Errorw(ctx, "failed to get clipboard", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	lastModified := clipboard.UpdatedAt.UTC().Format(http.TimeFormat)
	if ifLastModified != "" && lastModified == ifLastModified {
		h.log.Debugw(ctx, "Not modified", "id", sid)
		rw.WriteHeader(http.StatusNotModified)
		return
	}

	h.log.Debugw(ctx, "Got session", "id", sid)
	rw.Header().Set(LastModifiedHeader, lastModified)
	rw.Header().Set(ContentTypeHeader, clipboard.ContentType)
	if _, err = rw.Write(clipboard.Content); err != nil {
		h.log.Errorw(ctx, "failed to write content", err)
	}
}

func (h *SessionHandler) SetClipboard(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx         = r.Context()
		contentType = r.Header.Get(ContentTypeHeader)
		sessionID   = chi.URLParam(r, "sessionID")
	)

	if strings.ToLower(contentType) != "text/plain" {
		h.log.Debugw(ctx, "Content-Type is not text/plain")
		h.resp.SendBadRequest(ctx, rw, "Content-Type text/plain is required")
		return
	}

	if sessionID == "" {
		h.log.Debugw(ctx, "sessionID is empty")
		h.resp.SendBadRequest(ctx, rw, "sessionID param is required")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Errorw(ctx, "failed to read body", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		h.log.Errorw(ctx, "failed to parse sessionID", err)
		h.resp.SendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value")
		return
	}

	clipboard, err := h.clipboardRepo.SetBySessionID(sid, contentType, body)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			h.log.Debugw(ctx, "session not found", "id", sessionID)
			h.resp.SendNotFound(ctx, rw, "Session with provided ID not found")
			return
		}

		h.log.Errorw(ctx, "failed to set content", err)
		h.resp.SendInternalServerError(ctx, rw)
		return
	}
	go func() {
		if err := h.service.UpdateUpdatedAt(ctx, sid); err != nil {
			h.log.Errorw(ctx, "failed to update session updated_at", err)
		}
	}()

	h.log.Debugw(ctx, "Set content", "id", sessionID)
	rw.Header().Set(LastModifiedHeader, clipboard.UpdatedAt.UTC().Format(http.TimeFormat))
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
