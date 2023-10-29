package session

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type Repository interface {
	Get(id string) (*dal.Session, error)
	Create() (*dal.Session, error)
}

type Handler struct {
	sessionRepo Repository

	log trace.Logger
}

func NewSessionHandler(sessionRepo Repository, log trace.Logger) *Handler {
	return &Handler{
		sessionRepo: sessionRepo,
		log:         log,
	}
}

func (s *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", s.Create)
	r.Get("/{id}", s.Get)
}

func (s *Handler) Create(rw http.ResponseWriter, r *http.Request) {
	var (
		session *dal.Session
		body    []byte
		err     error
	)

	if session, err = s.sessionRepo.Create(); err != nil {
		s.log.Errorw(r.Context(), "failed to create session", err)
		rest.SendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(r.Context(), "Created session", "id", session.SessionID)
	if body, err = rest.ToJSON(session); err != nil {
		s.log.Errorw(r.Context(), "failed to marshal session", err)
		rest.SendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	rest.Send(r.Context(), rw, http.StatusCreated, rest.ContentTypeJSON, body, s.log)
}

func (s *Handler) Get(rw http.ResponseWriter, r *http.Request) {
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
			rest.SendNotFound(r.Context(), rw, s.log)
			return
		}

		s.log.Errorw(r.Context(), "failed to get session", err)
		rest.SendInternalServerError(r.Context(), rw, s.log)
		return
	}

	s.log.Debugw(r.Context(), "Got session", "id", session.SessionID)
	if body, err = rest.ToJSON(session); err != nil {
		s.log.Errorw(r.Context(), "failed to marshal session", err)
		rest.SendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	rest.Send(r.Context(), rw, http.StatusOK, rest.ContentTypeJSON, body, s.log)
}
