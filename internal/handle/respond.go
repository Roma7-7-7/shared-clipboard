package handle

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/log"
)

const (
	ContentTypeHeader     = "Content-Type"
	ContentTypeJSON       = "application/json"
	LastModifiedHeader    = "Last-Modified"
	IfModifiedSinceHeader = "If-Modified-Since"
)

type genericErrorResponse struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type responder struct {
	log log.TracedLogger
}

func (r *responder) Send(ctx context.Context, rw http.ResponseWriter, status int, headers map[string][]string, value interface{}) {
	body, err := json.Marshal(value)
	if err != nil {
		r.log.Errorw(ctx, "Failed to marshal response", err)
		r.SendInternalServerError(ctx, rw)
		return
	}

	contentTypeSet := false
	for k, v := range headers {
		if strings.EqualFold(k, ContentTypeHeader) {
			contentTypeSet = true
		}
		for _, v := range v {
			rw.Header().Add(k, v)
		}
	}
	if !contentTypeSet {
		rw.Header().Set(ContentTypeJSON, ContentTypeJSON)
	}
	rw.WriteHeader(status)

	if n, err := rw.Write(body); err != nil {
		// no reason to return error if we already wrote some bytes
		r.log.Errorw(ctx, "Failed to write response", "bytesWritten", n, err)
	}
}

func (r *responder) SendError(ctx context.Context, rw http.ResponseWriter, status int, code, message string, details any) {
	r.Send(ctx, rw, status, nil, genericErrorResponse{
		Error:   true,
		Code:    code,
		Message: message,
		Details: details,
	})
}

func (r *responder) SendUnauthorized(ctx context.Context, rw http.ResponseWriter) {
	r.Send(ctx, rw, http.StatusUnauthorized, nil, genericErrorResponse{
		Error:   true,
		Code:    domain.ErrorCodeUnauthorized.Value,
		Message: "Request is not authorized",
	})
}

func (r *responder) SendBadRequest(ctx context.Context, rw http.ResponseWriter, message string) {
	r.Send(ctx, rw, http.StatusBadRequest, nil, genericErrorResponse{
		Error:   true,
		Code:    domain.ErrorBadRequest.Value,
		Message: message,
	})
}

func (r *responder) SendNotFound(ctx context.Context, rw http.ResponseWriter, message string) {
	r.Send(ctx, rw, http.StatusNotFound, nil, genericErrorResponse{
		Error:   true,
		Code:    domain.ErrorCodeNotFound.Value,
		Message: message,
	})
}

func (r *responder) SendInternalServerError(ctx context.Context, rw http.ResponseWriter) {
	r.Send(ctx, rw, http.StatusInternalServerError, nil, genericErrorResponse{
		Error:   true,
		Code:    domain.ErrorCodeInternalServerError.Value,
		Message: "Internal server error",
	})
}
