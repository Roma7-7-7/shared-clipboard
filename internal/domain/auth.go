package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Roma7-7-7/shared-clipboard/internal/log"
)

type (
	JTIService struct {
		client RedisClient
		log    log.TracedLogger
	}
)

func NewJTIService(client RedisClient, log log.TracedLogger) *JTIService {
	return &JTIService{
		client: client,
		log:    log,
	}
}

func (s *JTIService) CreateBlockedJTI(ctx context.Context, jti string, expires time.Time) error {
	key := jtiKey(jti)
	s.log.Infow(ctx, "Creating blocked JTI", "key", key, "expires", expires)
	set := s.client.Set(ctx, key, "blocked", expires.Sub(time.Now()))
	if set.Err() != nil {
		return fmt.Errorf("set blocked jti with key=%q: %w", key, set.Err())
	}
	return nil
}

func (s *JTIService) IsBlockedJTIExists(ctx context.Context, jti string) (bool, error) {
	key := jtiKey(jti)
	s.log.Infow(ctx, "Checking blocked JTI", "key", key)
	get := s.client.Get(ctx, key)
	if get.Err() != nil {
		if errors.Is(get.Err(), redis.Nil) {
			return false, nil
		}

		return false, fmt.Errorf("get blocked jti with key=%q: %w", key, get.Err())
	}
	return true, nil
}

func jtiKey(jti string) string {
	return fmt.Sprintf("jti:%s", jti)
}
