package cookie

import (
	"fmt"
	"net/http"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
)

const (
	accessTokenCookieName = "accessToken"
)

type (
	JWTProcessor interface {
		ToAccessToken(userID uint64, name string) (string, error)
	}

	Processor struct {
		jwtProcessor JWTProcessor

		path   string
		domain string
	}
)

func NewProcessor(jwtProcessor JWTProcessor, conf config.Cookie) *Processor {
	return &Processor{
		jwtProcessor: jwtProcessor,

		path:   conf.Path,
		domain: conf.Domain,
	}
}

func (p *Processor) ToAccessToken(id uint64, name string) (*http.Cookie, error) {
	var (
		res = &http.Cookie{Name: accessTokenCookieName, Path: p.path, Domain: p.domain}
		err error
	)

	if res.Value, err = p.jwtProcessor.ToAccessToken(id, name); err != nil {
		return res, fmt.Errorf("create access token: %w", err)
	}

	return res, nil
}
