package handle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		GetByID(ctx context.Context, id uint64) (*domain.Session, error)
		GetByUserID(ctx context.Context, userID uint64) ([]*domain.Session, error)
		Create(ctx context.Context, name string, userID uint64) (*domain.Session, error)
		Update(ctx context.Context, sessionID, userID uint64, name string) (*domain.Session, error)
		Delete(ctx context.Context, sessionID, userID uint64) error
	}

	ClipboardRepository interface {
		GetBySessionID(id uint64) (*dal.Clipboard, error)
		SetBySessionID(id uint64, contentType string, content []byte) (*dal.Clipboard, error)
	}

	SessionHandler struct {
		service       SessionService
		clipboardRepo ClipboardRepository
		log           log.TracedLogger
	}
)

func NewSessionHandler(sessionService SessionService, clipboardRepo ClipboardRepository, log log.TracedLogger) *SessionHandler {
	return &SessionHandler{
		service:       sessionService,
		clipboardRepo: clipboardRepo,
		log:           log,
	}
}

func (s *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", s.Create)
	r.Get("/", s.GetAllByUserID)
	r.Get("/{sessionID}", s.GetByID)
	r.Put("/{sessionID}", s.Update)
	r.Delete("/{sessionID}", s.Delete)
	r.Get("/{sessionID}/clipboard", s.GetClipboard)
	r.Put("/{sessionID}/clipboard", s.SetClipboard)
}

func (s *SessionHandler) GetByID(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		tid       = domain.TraceIDFromContext(ctx)
		sessionID = chi.URLParam(r, "sessionID")
	)
	s.log.Debugw(tid, "Get session by ID", "sessionID", sessionID)

	if sessionID == "" {
		s.log.Debugw(tid, "sessionID is empty")
		sendBadRequest(ctx, rw, "sessionID param is required", s.log)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		s.log.Errorw(tid, "failed to parse sessionID", err)
		sendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value", s.log)
		return
	}

	session, err := s.service.GetByID(ctx, sid)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			s.log.Debugw(tid, "session not found", "sessionID", sessionID)
			sendNotFound(ctx, rw, "Session with provided ID not found", s.log)
			return
		}

		s.log.Errorw(tid, "failed to get session", "sessionID", sessionID, err)
		sendInternalServerError(ctx, rw, s.log)
		return
	}

	s.log.Debugw(tid, "Got session", "sessionID", session.ID)
	body, err := rest.ToJSON(toDTO(session))
	if err != nil {
		s.log.Errorw(tid, "failed to marshal session", err)
		sendErrorMarshalBody(ctx, rw, s.log)
		return
	}

	rw.Header().Set(rest.LastModifiedHeader, session.UpdatedAt.Format(http.TimeFormat))
	rest.Send(ctx, rw, http.StatusOK, rest.ContentTypeJSON, body, s.log)
}

func (s *SessionHandler) GetAllByUserID(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tid = domain.TraceIDFromContext(ctx)
	)

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		s.log.Debugw(tid, "user not found in context")
		sendUnauthorized(ctx, rw, s.log)
		return
	}
	s.log.Debugw(tid, "Get all sessions by user", "userID", auth.UserID)

	sessions, err := s.service.GetByUserID(ctx, auth.UserID)
	if err != nil {
		s.log.Errorw(tid, "failed to get sessions", "userID", auth.UserID, err)
		sendInternalServerError(ctx, rw, s.log)
		return
	}

	s.log.Debugw(tid, "Got sessions", "count", len(sessions))
	res := make([]*Session, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, toDTO(session))
	}

	body, err := rest.ToJSON(res)
	if err != nil {
		s.log.Errorw(tid, "failed to marshal sessions", err)
		sendErrorMarshalBody(ctx, rw, s.log)
		return
	}

	rest.Send(ctx, rw, http.StatusOK, rest.ContentTypeJSON, body, s.log)
}

func (s *SessionHandler) Create(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		tid = domain.TraceIDFromContext(ctx)
	)

	user, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		s.log.Debugw(tid, "user not found in context")
		sendUnauthorized(ctx, rw, s.log)
		return
	}

	var req sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Debugw(tid, "failed to decode body", err)
		sendBadRequest(ctx, rw, "failed to parse request", s.log)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		s.log.Debugw(tid, "name is empty")
		sendBadRequest(ctx, rw, "name param is required", s.log)
		return
	}

	session, err := s.service.Create(ctx, req.Name, user.UserID)
	if err != nil {
		s.log.Errorw(tid, "failed to create session", err)
		sendInternalServerError(ctx, rw, s.log)
		return
	}
	s.log.Debugw(tid, "Created session", "id", session.ID)

	body, err := rest.ToJSON(toDTO(session))
	if err != nil {
		s.log.Errorw(tid, "failed to marshal session", err)
		sendErrorMarshalBody(ctx, rw, s.log)
		return
	}

	rw.Header().Set(rest.LastModifiedHeader, session.UpdatedAt.Format(http.TimeFormat))
	rest.Send(ctx, rw, http.StatusCreated, rest.ContentTypeJSON, body, s.log)
}

func (s *SessionHandler) Update(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		sessionID = chi.URLParam(r, "sessionID")
		tid       = domain.TraceIDFromContext(ctx)
	)

	if sessionID == "" {
		s.log.Debugw(tid, "sessionID is empty")
		sendBadRequest(ctx, rw, "sessionID param is required", s.log)
		return
	}

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		s.log.Debugw(tid, "user not found in context")
		sendUnauthorized(ctx, rw, s.log)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		s.log.Errorw(tid, "failed to parse sessionID", err)
		sendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value", s.log)
		return
	}

	var req sessionRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Debugw(tid, "failed to decode body", err)
		sendBadRequest(ctx, rw, "failed to parse request", s.log)
		return
	}

	session, err := s.service.Update(ctx, sid, auth.UserID, req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			s.log.Debugw(tid, "session not found", "sessionID", sessionID)
			sendNotFound(ctx, rw, "Session with provided ID not found", s.log)
			return
		}

		if errors.Is(err, domain.ErrSessionPermissionDenied) {
			s.log.Debugw(tid, "permission denied", "sessionID", sessionID)
			sendForbidden(ctx, rw, "Permission denied", s.log)
			return
		}

		s.log.Errorw(tid, "failed to update session", err)
		sendInternalServerError(ctx, rw, s.log)
		return
	}

	body, err := rest.ToJSON(toDTO(session))
	if err != nil {
		s.log.Errorw(tid, "failed to marshal session", err)
		sendErrorMarshalBody(ctx, rw, s.log)
		return
	}

	rw.Header().Set(rest.LastModifiedHeader, session.UpdatedAt.Format(http.TimeFormat))
	rest.Send(ctx, rw, http.StatusOK, rest.ContentTypeJSON, body, s.log)
}

func (s *SessionHandler) Delete(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx       = r.Context()
		sessionID = chi.URLParam(r, "sessionID")
		tid       = domain.TraceIDFromContext(ctx)
	)

	if sessionID == "" {
		s.log.Debugw(tid, "sessionID is empty")
		sendBadRequest(ctx, rw, "sessionID param is required", s.log)
		return
	}

	auth, ok := domain.AuthorityFromContext(ctx)
	if !ok {
		s.log.Debugw(tid, "user not found in context")
		sendUnauthorized(ctx, rw, s.log)
		return
	}

	sid, err := strconv.ParseUint(sessionID, 10, 64)
	if err != nil {
		s.log.Errorw(tid, "failed to parse sessionID", err)
		sendBadRequest(ctx, rw, "sessionID param must be a valid uint64 value", s.log)
		return
	}

	if err = s.service.Delete(ctx, sid, auth.UserID); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			s.log.Debugw(tid, "session not found", "sessionID", sessionID)
			sendNotFound(ctx, rw, "Session with provided ID not found", s.log)
			return
		}

		if errors.Is(err, domain.ErrSessionPermissionDenied) {
			s.log.Debugw(tid, "permission denied", "sessionID", sessionID)
			sendForbidden(ctx, rw, "Permission denied", s.log)
			return
		}

		s.log.Errorw(tid, "failed to delete session", err)
		sendInternalServerError(ctx, rw, s.log)
		return
	}

	rest.SendNoContent(ctx, rw, s.log)
}

func (s *SessionHandler) GetClipboard(rw http.ResponseWriter, r *http.Request) {
	var (
		ifLastModified = r.Header.Get(rest.IfModifiedSinceHeader)
		sessionID      string
		sid            uint64
		clipboard      *dal.Clipboard
		err            error
	)

	sessionID = chi.URLParam(r, "sessionID")
	if sessionID == "" {
		s.log.Debugw(domain.TraceIDFromContext(r.Context()), "sessionID is empty")
		sendBadRequest(r.Context(), rw, "sessionID param is required", s.log)
		return
	}

	if sid, err = strconv.ParseUint(sessionID, 10, 64); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to parse sessionID", err)
		sendBadRequest(r.Context(), rw, "sessionID param must be a valid uint64 value", s.log)
		return
	}

	if clipboard, err = s.clipboardRepo.GetBySessionID(sid); err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(domain.TraceIDFromContext(r.Context()), "clipboard not found", "id", sessionID)
			rest.SendNoContent(r.Context(), rw, s.log)
			return
		}

		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to get clipboard", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	lastModified := clipboard.UpdatedAt.UTC().Format(http.TimeFormat)
	if ifLastModified != "" && lastModified == ifLastModified {
		s.log.Debugw(domain.TraceIDFromContext(r.Context()), "Not modified", "id", sid)
		rw.WriteHeader(http.StatusNotModified)
		return
	}

	s.log.Debugw(domain.TraceIDFromContext(r.Context()), "Got session", "id", sid)
	rw.Header().Set(rest.LastModifiedHeader, lastModified)
	rw.Header().Set(rest.ContentTypeHeader, clipboard.ContentType)
	if _, err = rw.Write(clipboard.Content); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to write content", err)
	}
}

func (s *SessionHandler) SetClipboard(rw http.ResponseWriter, r *http.Request) {
	var (
		contentType = r.Header.Get(rest.ContentTypeHeader)
		sessionID   string
		sid         uint64
		clipboard   *dal.Clipboard
		body        []byte
		err         error
	)

	if contentType != "text/plain" {
		s.log.Debugw(domain.TraceIDFromContext(r.Context()), "Content-Type is not text/plain")
		sendBadRequest(r.Context(), rw, fmt.Sprintf("Content-Type %s is not supported", contentType), s.log)
		return
	}

	sessionID = chi.URLParam(r, "sessionID")
	if sessionID == "" {
		s.log.Debugw(domain.TraceIDFromContext(r.Context()), "sessionID is empty")
		sendBadRequest(r.Context(), rw, "sessionID param is required", s.log)
		return
	}

	if body, err = io.ReadAll(r.Body); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to read body", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	if sid, err = strconv.ParseUint(sessionID, 10, 64); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to parse sessionID", err)
		sendBadRequest(r.Context(), rw, "sessionID param must be a valid uint64 value", s.log)
		return
	}

	if clipboard, err = s.clipboardRepo.SetBySessionID(sid, contentType, body); err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(domain.TraceIDFromContext(r.Context()), "session not found", "id", sessionID)
			sendNotFound(r.Context(), rw, "Session with provided ID not found", s.log)
			return
		}

		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to set content", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(domain.TraceIDFromContext(r.Context()), "Set content", "id", sessionID)
	rw.Header().Set(rest.LastModifiedHeader, clipboard.UpdatedAt.UTC().Format(http.TimeFormat))
	rest.SendNoContent(r.Context(), rw, s.log)
}

func toDTO(session *domain.Session) *Session {
	return &Session{
		SessionID:       session.ID,
		Name:            session.Name,
		CreatedAtMillis: session.CreatedAt.UnixMilli(),
		UpdatedAtMillis: session.UpdatedAt.UnixMilli(),
	}
}
