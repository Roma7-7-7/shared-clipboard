package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type errorCode string

const errorResponseTmpl = `{"code": "%s", "message": "%s"}`

var (
	errorCodeInternalServerError = errorCode("ERR_0500")
	errorCodeNotFound            = errorCode("ERR_0404")
	errorCodeMethodNotAllowed    = errorCode("ERR_0405")

	errorCodeMarshalResponse = errorCode("ERR_1000")
)

func notFoundErrorBody() []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, errorCodeNotFound, "Not Found"))
}

func methodNotAllowedErrorBody(method string) []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, errorCodeMethodNotAllowed, fmt.Sprintf("Method %s is not allowed", method)))
}

func internalServerErrorBody() []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, errorCodeInternalServerError, "Internal Server Error"))
}

func marshalErrorBody() []byte {
	return []byte(fmt.Sprintf(errorResponseTmpl, errorCodeMarshalResponse, "Failed to marshal response"))
}

func errorResponse(errorCode errorCode, message string, details any) ([]byte, error) {
	body := map[string]interface{}{
		"code":    string(errorCode),
		"message": message,
	}
	if details != nil {
		body["details"] = details
	}

	return handler.ToJSON(body)
}

func sendErrorMarshalBody(ctx context.Context, rw http.ResponseWriter, log trace.Logger) {
	handler.Send(ctx, rw, http.StatusInternalServerError, handler.ContentTypeJSON, marshalErrorBody(), log)
}

func sendInternalServerError(ctx context.Context, rw http.ResponseWriter, log trace.Logger) {
	handler.Send(ctx, rw, http.StatusInternalServerError, handler.ContentTypeJSON, internalServerErrorBody(), log)
}
