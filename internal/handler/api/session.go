package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type SessionRepository interface {
	Get(id string) (*dal.Session, error)
	Create() (*dal.Session, error)
}

type SessionHandler struct {
	sessionRepo SessionRepository

	log log.TracedLogger
}

func NewSessionHandler(sessionRepo SessionRepository, log log.TracedLogger) *SessionHandler {
	return &SessionHandler{
		sessionRepo: sessionRepo,
		log:         log,
	}
}

func (s *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", s.Create)
	r.Get("/{id}", s.Get)
}

func (s *SessionHandler) Create(rw http.ResponseWriter, r *http.Request) {
	var (
		session *dal.Session
		body    []byte
		err     error
	)

	if session, err = s.sessionRepo.Create(); err != nil {
		s.log.Errorw(trace.ID(r.Context()), "failed to create session", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(trace.ID(r.Context()), "Created session", "id", session.SessionID)
	if body, err = rest.ToJSON(session); err != nil {
		s.log.Errorw(trace.ID(r.Context()), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	rest.Send(r.Context(), rw, http.StatusCreated, rest.ContentTypeJSON, body, s.log)
}

func (s *SessionHandler) Get(rw http.ResponseWriter, r *http.Request) {
	var (
		sessionID string
		session   *dal.Session
		body      []byte
		err       error
	)

	sessionID = chi.URLParam(r, "id")

	if session, err = s.sessionRepo.Get(sessionID); err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(trace.ID(r.Context()), "session not found", "id", sessionID)
			sendNotFound(r.Context(), rw, s.log)
			return
		}

		s.log.Errorw(trace.ID(r.Context()), "failed to get session", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(trace.ID(r.Context()), "Got session", "id", session.SessionID)
	if body, err = rest.ToJSON(session); err != nil {
		s.log.Errorw(trace.ID(r.Context()), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	rest.Send(r.Context(), rw, http.StatusOK, rest.ContentTypeJSON, body, s.log)
}
