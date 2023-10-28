package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type Response struct {
	Error   bool   `json:"error,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type Error struct {
	RootError  error `json:"-"`
	HTTPStatus int   `json:"-"`
	Response
}

func NewAPIError(err error, status int, code errorCode, message string, details any) Error {
	return Error{
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

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.RootError)
}

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

func (a *Service) Create(c echo.Context) error {
	session, err := a.sessionRepo.Create()
	if err != nil {
		return NewAPIError(err, http.StatusInternalServerError, errorCodeCreateSession, "failed to create session", nil)
	}
	a.log.Debugw(c.Request().Context(), "Created session: ", "id", session.ID)
	return c.JSON(http.StatusCreated, session)
}
