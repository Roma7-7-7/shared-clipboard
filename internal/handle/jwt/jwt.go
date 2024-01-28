package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
)

type (
	Processor struct {
		issuer          string
		audience        []string
		expireInMinutes uint64

		secret []byte
	}

	Claims struct {
		Username string `json:"username"`
		jwt.RegisteredClaims
	}
)

func NewProcessor(conf config.JWT) *Processor {
	return &Processor{
		issuer:          conf.Issuer,
		audience:        conf.Audience,
		expireInMinutes: conf.ExpireInMinutes,

		secret: []byte(conf.Secret),
	}
}

func (p *Processor) ToAccessToken(userID uint64, name string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Username: name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  p.audience,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * time.Duration(p.expireInMinutes))),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	})

	signedString, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedString, nil
}

func (p *Processor) ParseAccessToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return p.secret, nil
	})
}
