package api

import (
	"net/http"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type SessionRepository interface {
	Create() (*dal.Session, error)
}

type Service struct {
	sessionRepo SessionRepository

	log trace.Logger
}

func NewAPIService(sessionRepo SessionRepository, log trace.Logger) *Service {
	return &Service{
		sessionRepo: sessionRepo,

		log: log,
	}
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

	s.log.Debugw(r.Context(), "Created session", "id", session.ID)
	if body, err = handler.ToJSON(session); err != nil {
		s.log.Errorw(r.Context(), "failed to marshal session", err)
		sendErrorMarshalBody(r.Context(), rw, s.log)
		return
	}

	handler.Send(r.Context(), rw, http.StatusCreated, handler.ContentTypeJSON, body, s.log)
}
