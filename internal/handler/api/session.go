package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type SessionRepository interface {
	Get(id string) (*dal.Session, error)
	Create() (*dal.Session, error)
}

type Service struct {
	sessionRepo SessionRepository

	log trace.Logger
}

func NewSessionHandler(sessionRepo SessionRepository, log trace.Logger) *Service {
	return &Service{
		sessionRepo: sessionRepo,

		log: log,
	}
}

func (s *Service) RegisterRoutes(r chi.Router) {
	r.Post("/", s.Create)
	r.Get("/{id}", s.Get)
}

func (s *Service) Create(rw http.ResponseWriter, r *http.Request) {
	var (
		session *dal.Session
		body    []byte
		err     error
	)

	if session, err = s.sessionRepo.Create(); err != nil {
		s.log.Errorw(r.Context(), "failed to create session", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(r.Context(), "Created session", "id", session.SessionID)
	if body, err = handler.ToJSON(session); err != nil {
		s.log.Errorw(r.Context(), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	handler.Send(r.Context(), rw, http.StatusCreated, handler.ContentTypeJSON, body, s.log)
}

func (s *Service) Get(rw http.ResponseWriter, r *http.Request) {
	var (
		sessionID string
		session   *dal.Session
		body      []byte
		err       error
	)

	sessionID = chi.URLParam(r, "id")

	if session, err = s.sessionRepo.Get(sessionID); err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(r.Context(), "session not found", "id", sessionID)
			sendNotFound(r.Context(), rw, s.log)
			return
		}

		s.log.Errorw(r.Context(), "failed to get session", err)
		sendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(r.Context(), "Got session", "id", session.SessionID)
	if body, err = handler.ToJSON(session); err != nil {
		s.log.Errorw(r.Context(), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	handler.Send(r.Context(), rw, http.StatusOK, handler.ContentTypeJSON, body, s.log)
}
