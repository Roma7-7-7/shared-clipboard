package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
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

func sendNotFound(ctx context.Context, rw http.ResponseWriter, log log.TracedLogger) {
	rest.Send(ctx, rw, http.StatusNotFound, rest.ContentTypeJSON, notFoundErrorBody(), log)
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
