package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	log "go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
)

type SessionRepository interface {
	Create() (*dal.Session, error)
}

type API struct {
	sessionRepo SessionRepository

	log *log.SugaredLogger
}

func NewAPI(sessionRepo SessionRepository, log *log.SugaredLogger) *API {
	return &API{
		sessionRepo: sessionRepo,

		log: log,
	}
}

func (a *API) Create(c echo.Context) error {
	session, err := a.sessionRepo.Create()
	if err != nil {
		return NewAPIError(err, errorCodeCreateSession, "failed to create session", nil)
	}
	a.log.Debugw("Created session: ", "id", session.ID)
	return c.JSON(http.StatusCreated, session)
}
