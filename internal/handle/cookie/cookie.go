package cookie

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
)

const (
	accessTokenCookieName = "accessToken"
)

var (
	ErrAccessTokenNotFound = fmt.Errorf("access token not found")
	ErrParseAccessToken    = fmt.Errorf("parse access token")
)

type (
	JWTProcessor interface {
		ToAccessToken(userID uint64, name string) (string, error)
		ParseAccessToken(tokenString string) (*jwt.Token, error)
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

func (p *Processor) ExpireAccessToken() *http.Cookie {
	return &http.Cookie{Name: accessTokenCookieName, Path: p.path, Domain: p.domain, Expires: time.Now()}
}

func (p *Processor) AccessTokenFromRequest(r *http.Request) (*jwt.Token, error) {
	cookie, err := r.Cookie(accessTokenCookieName)
	if err != nil {
		return nil, ErrAccessTokenNotFound
	}

	token, err := p.jwtProcessor.ParseAccessToken(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParseAccessToken, err)
	}
	return token, err
}
