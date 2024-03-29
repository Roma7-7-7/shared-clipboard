package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/log"
)

var (
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionPermissionDenied = errors.New("session permission denied")
)

type (
	SessionFilter struct {
		Limit      int `validate:"required,gte=1,lte=100"`
		Name       string
		SortBy     string `validate:"omitempty,oneof=name updated_at"`
		SortByDesc bool
		Offset     int `validate:"gte=0"`
	}

	Session struct {
		ID        uint64
		Name      string
		UserID    uint64
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	SessionRepository interface {
		GetByID(id uint64) (*dal.Session, error)
		GetAllByUserID(userID uint64) ([]*dal.Session, error)
		FilterBy(dal.SessionFilter) ([]*dal.Session, int, error)
		Create(name string, userID uint64) (*dal.Session, error)
		Update(id uint64, name string) (*dal.Session, error)
		UpdateUpdatedAt(id uint64) error
		Delete(id uint64) error
	}

	SessionService struct {
		sessionRepo SessionRepository

		log log.TracedLogger
	}
)

func NewSessionService(sessionRepo SessionRepository, log log.TracedLogger) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		log:         log,
	}
}

func (s *SessionService) GetByID(ctx context.Context, userID, id uint64) (*Session, error) {
	s.log.Debugw(ctx, "get session by id", "sessionID", id)

	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(ctx, "session not found")
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf("get session by id=%d: %w", id, err)
	}
	if session.UserID != userID {
		return nil, ErrSessionPermissionDenied
	}

	s.log.Debugw(ctx, "session found", "session", session)
	return toSession(session), nil
}

func (s *SessionService) GetByUserID(ctx context.Context, userID uint64) ([]*Session, error) {
	s.log.Debugw(ctx, "get sessions by userID", "userID", userID)

	sessions, err := s.sessionRepo.GetAllByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("get sessions by userID=%d: %w", userID, err)
	}

	s.log.Debugw(ctx, "sessions found", "count", len(sessions))
	res := make([]*Session, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, toSession(session))
	}
	return res, nil
}

func (s *SessionService) FilterBy(ctx context.Context, userID uint64, filter SessionFilter) ([]*Session, int, error) {
	s.log.Debugw(ctx, "filter sessions", "userID", userID, "filter", filter)

	filterOpts := make([]func(dal.SessionFilter) dal.SessionFilter, 0, 3)
	if filter.Name != "" {
		filterOpts = append(filterOpts, dal.WithName(filter.Name))
	}
	switch filter.SortBy {
	case "name":
		filterOpts = append(filterOpts, dal.WithSortByName(filter.SortByDesc))
	case "updated_at":
		filterOpts = append(filterOpts, dal.WithSortByUpdateAt(filter.SortByDesc))
	}
	filterOpts = append(filterOpts, dal.WithOffset(filter.Offset))
	dalFilter := dal.NewSessionFilter(userID, filter.Limit, filterOpts...)

	sessions, total, err := s.sessionRepo.FilterBy(dalFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("filter sessions by %v: %w", filter, err)
	}

	s.log.Debugw(ctx, "sessions found", "count", len(sessions), "total", total)
	res := make([]*Session, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, toSession(session))
	}
	return res, total, nil
}

func (s *SessionService) Create(ctx context.Context, userID uint64, name string) (*Session, error) {
	s.log.Debugw(ctx, "create session", "name", name, "userID", userID)

	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is empty")
	}

	session, err := s.sessionRepo.Create(name, userID)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	s.log.Debugw(ctx, "session created", "session", session)
	return toSession(session), nil
}

func (s *SessionService) Update(ctx context.Context, userID, sessionID uint64, name string) (*Session, error) {
	s.log.Debugw(ctx, "update session", "sessionID", sessionID, "name", name)

	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is empty")
	}

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf("get session by id=%d: %w", sessionID, err)
	}

	if session.UserID != userID {
		return nil, ErrSessionPermissionDenied
	}

	updated, err := s.sessionRepo.Update(sessionID, name)
	if err != nil {
		return nil, fmt.Errorf("update session by id=%q: %w", sessionID, err)
	}

	s.log.Debugw(ctx, "session updated", "session", updated)
	return toSession(updated), nil
}

func (s *SessionService) UpdateUpdatedAt(ctx context.Context, sessionID uint64) error {
	s.log.Debugw(ctx, "update session updated_at", "sessionID", sessionID)

	if err := s.sessionRepo.UpdateUpdatedAt(sessionID); err != nil {
		return fmt.Errorf("update session updated_at by id=%d: %w", sessionID, err)
	}

	s.log.Debugw(ctx, "session updated_at updated", "sessionID", sessionID)
	return nil
}

func (s *SessionService) Delete(ctx context.Context, userID, sessionID uint64) error {
	s.log.Debugw(ctx, "delete session", "sessionID", sessionID)

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			return ErrSessionNotFound
		}

		return fmt.Errorf("get session by id=%d: %w", sessionID, err)
	}

	if session.UserID != userID {
		return ErrSessionPermissionDenied
	}

	if err = s.sessionRepo.Delete(sessionID); err != nil {
		return fmt.Errorf("delete session by id=%d: %w", sessionID, err)
	}

	s.log.Debugw(ctx, "session deleted", "sessionID", sessionID)
	return nil
}

func toSession(session *dal.Session) *Session {
	return &Session{
		ID:        session.ID,
		Name:      session.Name,
		UserID:    session.UserID,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
}
