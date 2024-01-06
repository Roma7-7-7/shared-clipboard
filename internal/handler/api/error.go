package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
)

const errorResponseTmpl = `{"error": true, "code": "%s", "message": "%s"}`

type genericErrorResponse struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func badRequestErrorBody(message string) []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, domain.ErrorBadRequest, message))
}

func notFoundErrorBody(message string) []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, domain.ErrorCodeNotFound, message))
}

func unauthorizedErrorBody(message string) []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, domain.ErrorCodeUnauthorized, message))
}

func forbiddenErrorBody(message string) []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, domain.ErrorCodeForbidden, message))
}

func methodNotAllowedErrorBody(method string) []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, domain.ErrorCodeMethodNotAllowed, fmt.Sprintf("Method %s is not allowed", method)))
}

func internalServerErrorBody() []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, domain.ErrorCodeInternalServerError, "Internal server error"))
}

func marshalErrorBody() []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, domain.ErrorCodeMarshalResponse, "Failed to marshal response"))
}

func sendBadRequest(ctx context.Context, rw http.ResponseWriter, message string, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusBadRequest, rest.ContentTypeJSON, badRequestErrorBody(message), log)
}

func sendNotFound(ctx context.Context, rw http.ResponseWriter, message string, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusNotFound, rest.ContentTypeJSON, notFoundErrorBody(message), log)
}

func sendUnauthorized(ctx context.Context, rw http.ResponseWriter, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusUnauthorized, rest.ContentTypeJSON, unauthorizedErrorBody("Request is not authorized"), log)
}

func sendForbidden(ctx context.Context, rw http.ResponseWriter, message string, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusUnauthorized, rest.ContentTypeJSON, forbiddenErrorBody(message), log)
}

func sendErrorMarshalBody(ctx context.Context, rw http.ResponseWriter, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusInternalServerError, rest.ContentTypeJSON, marshalErrorBody(), log)
}

func sendErrorMethodNotAllowed(ctx context.Context, method string, rw http.ResponseWriter, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusMethodNotAllowed, rest.ContentTypeJSON, methodNotAllowedErrorBody(method), log)
}

func sendInternalServerError(ctx context.Context, rw http.ResponseWriter, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusInternalServerError, rest.ContentTypeJSON, internalServerErrorBody(), log)
}

func sendRenderableError(ctx context.Context, err *domain.RenderableError, rw http.ResponseWriter, log log.TracedLogger) {
	bytes, mErr := json.Marshal(genericErrorResponse{
		Error:   true,
		Code:    err.Code.Value,
		Message: err.Message,
		Details: err.Details,
	})
	if mErr != nil {
		log.Errorw(domain.TraceIDFromContext(ctx), "failed to marshal renderable error", mErr)
		sendErrorMarshalBody(ctx, rw, log)
		return
	}

	rest.Send(ctx, rw, err.Code.StatusCode, rest.ContentTypeJSON, bytes, log)
}
