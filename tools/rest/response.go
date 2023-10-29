package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

const (
	ContentTypeHeader = "Content-Type"

	ContentTypeJSON       = "application/json"
	ContentTypeJavaScript = "application/javascript"
)

func ToJSON(data any) ([]byte, error) {
	marshal, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	return marshal, err
}

func Send(ctx context.Context, rw http.ResponseWriter, status int, contentType string, body []byte, log trace.Logger) {
	rw.Header().Set(ContentTypeHeader, contentType)
	rw.WriteHeader(status)
	if _, err := rw.Write(body); err != nil {
		log.Errorw(ctx, "Failed to write response", err)
	}
}
