package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Roma7-7-7/shared-clipboard/internal/log"
)

type (
	Clipboard struct {
		SessionID   uint64
		ContentType string
		Content     []byte
		UpdatedAt   time.Time
	}

	ClipboardService struct {
		client RedisClient
		log    log.TracedLogger
	}
)

func NewClipboardService(client RedisClient, log log.TracedLogger) *ClipboardService {
	return &ClipboardService{
		client: client,
		log:    log,
	}
}

func (s *ClipboardService) GetBySessionID(ctx context.Context, id uint64) (*Clipboard, error) {
	key := clipboardKey(id)
	s.log.Debugw(ctx, "Getting clipboard", "key", key)

	cmd := s.client.Get(ctx, key)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			s.log.Debugw(ctx, "Clipboard not found", "key", key)
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get clipboard with key=%q: %w", key, cmd.Err())
	}
	bytes, err := cmd.Bytes()
	if err != nil {
		return nil, fmt.Errorf("get clipboard bytes with key=%q: %w", key, err)
	}
	var clipboard Clipboard
	if err = json.Unmarshal(bytes, &clipboard); err != nil {
		return nil, fmt.Errorf("unmarshal clipboard with key=%q: %w", key, err)
	}

	return &clipboard, nil
}

func (s *ClipboardService) SetBySessionID(ctx context.Context, id uint64, contentType string, content []byte) (*Clipboard, error) {
	key := clipboardKey(id)
	s.log.Debugw(ctx, "Setting clipboard", "key", key)
	clipboard := &Clipboard{
		SessionID:   id,
		ContentType: contentType,
		Content:     content,
		UpdatedAt:   time.Now(),
	}
	bytes, err := json.Marshal(clipboard)
	if err != nil {
		return nil, fmt.Errorf("marshal clipboard: %w", err)
	}

	cmd := s.client.Set(ctx, key, bytes, 24*time.Hour)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("set clipboard with key=%q: %w", key, cmd.Err())
	}

	return clipboard, nil
}

func clipboardKey(id uint64) string {
	return fmt.Sprintf("clipboard:%s", strconv.FormatUint(id, 10))
}
