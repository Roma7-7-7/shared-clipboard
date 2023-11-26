package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
)

type Session struct {
	SessionID uint64 `json:"session_id"`
	JoinKey   string `json:"join_key"`
	UpdatedAt int64  `json:"updated_at"`
}

type SessionRepository interface {
	GetByID(id uint64) (*dal.Session, error)
	GetByJoinKey(key string) (*dal.Session, error)
	Create() (*dal.Session, error)
}

type ClipboardRepository interface {
	GetBySessionID(id uint64) (*dal.Clipboard, error)
	SetBySessionID(id uint64, contentType string, content []byte) (*dal.Clipboard, error)
}

type SessionHandler struct {
	sessionRepo   SessionRepository
	clipboardRepo ClipboardRepository

	log log.TracedLogger
}

func NewSessionHandler(sessionRepo SessionRepository, clipboardRepo ClipboardRepository, log log.TracedLogger) *SessionHandler {
	return &SessionHandler{
		sessionRepo:   sessionRepo,
		clipboardRepo: clipboardRepo,
		log:           log,
	}
}

func (s *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", s.Create)
	r.Get("/", s.GetByJoinKey)
	r.Get("/{sessionID}", s.GetByID)
	r.Get("/{sessionID}/clipboard", s.GetClipboard)
	r.Put("/{sessionID}/clipboard", s.SetClipboard)
}

func (s *SessionHandler) GetByID(rw http.ResponseWriter, r *http.Request) {
	var (
		sessionID string
		sid       uint64
		session   *dal.Session
		body      []byte
		err       error
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

	if session, err = s.sessionRepo.GetByID(sid); err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(domain.TraceIDFromContext(r.Context()), "session not found", "id", sessionID)
			sendNotFound(r.Context(), rw, "Session with provided ID not found", s.log)
			return
		}

		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to get session", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(domain.TraceIDFromContext(r.Context()), "Got session", "id", session.SessionID)
	if body, err = rest.ToJSON(toDTO(session)); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	rw.Header().Set(rest.LastModifiedHeader, session.UpdatedAt.Format(http.TimeFormat))
	rest.Send(r.Context(), rw, http.StatusOK, rest.ContentTypeJSON, body, s.log)
}

func (s *SessionHandler) GetByJoinKey(rw http.ResponseWriter, r *http.Request) {
	var (
		joinKey string
		session *dal.Session
		body    []byte
		err     error
	)

	joinKey = r.URL.Query().Get("joinKey")
	if joinKey == "" {
		s.log.Debugw(domain.TraceIDFromContext(r.Context()), "joinKey is empty")
		sendBadRequest(r.Context(), rw, "joinKey param is required", s.log)
		return
	}

	if session, err = s.sessionRepo.GetByJoinKey(joinKey); err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(domain.TraceIDFromContext(r.Context()), "session not found", "id", joinKey)
			sendNotFound(r.Context(), rw, "Session with provided join key not found", s.log)
			return
		}

		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to get session", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(domain.TraceIDFromContext(r.Context()), "Got session", "id", session.SessionID)
	if body, err = rest.ToJSON(toDTO(session)); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	rw.Header().Set(rest.LastModifiedHeader, session.UpdatedAt.Format(http.TimeFormat))
	rest.Send(r.Context(), rw, http.StatusOK, rest.ContentTypeJSON, body, s.log)
}

func (s *SessionHandler) Create(rw http.ResponseWriter, r *http.Request) {
	var (
		session *dal.Session
		body    []byte
		err     error
	)

	if session, err = s.sessionRepo.Create(); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to create session", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(domain.TraceIDFromContext(r.Context()), "Created session", "id", session.SessionID)
	if body, err = rest.ToJSON(toDTO(session)); err != nil {
		s.log.Errorw(domain.TraceIDFromContext(r.Context()), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	rw.Header().Set(rest.LastModifiedHeader, session.UpdatedAt.Format(http.TimeFormat))
	rest.Send(r.Context(), rw, http.StatusCreated, rest.ContentTypeJSON, body, s.log)
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

func toDTO(session *dal.Session) *Session {
	return &Session{
		SessionID: session.SessionID,
		JoinKey:   session.JoinKey,
		UpdatedAt: session.UpdatedAt.UnixMilli(),
	}
}
