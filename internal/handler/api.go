package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	log "go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
)

type Response struct {
	Error   bool   `json:"error,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type APIError struct {
	RootError  error `json:"-"`
	HTTPStatus int   `json:"-"`
	Response
}

func NewAPIError(err error, status int, code errorCode, message string, details any) APIError {
	return APIError{
		RootError:  err,
		HTTPStatus: status,
		Response: Response{
			Error:   true,
			Code:    string(code),
			Message: message,
			Details: details,
		},
	}
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.RootError)
}

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
		return NewAPIError(err, http.StatusInternalServerError, errorCodeCreateSession, "failed to create session", nil)
	}
	a.log.Debugw("Created session: ", "id", session.ID)
	return c.JSON(http.StatusCreated, session)
}
