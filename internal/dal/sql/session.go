package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) (*SessionRepository, error) {
	return &SessionRepository{
		db: db,
	}, nil
}

func (r *SessionRepository) GetByID(id uint64) (*dal.Session, error) {
	var res dal.Session

	if err := r.db.QueryRow("SELECT session_id, user_id, created_at, updated_at FROM sessions WHERE session_id = $1", id).
		Scan(
			&res.SessionID,
			&res.UserID,
			&res.CreatedAt,
			&res.UpdatedAt,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session with session_id=%d not found: %w", id, dal.ErrNotFound)
		}

		return nil, fmt.Errorf("get session by session_id=%d: %w", id, err)
	}

	return &res, nil
}

func (r *SessionRepository) GetAllByUserID(userID uint64) ([]*dal.Session, error) {
	res := make([]*dal.Session, 0, 10)

	rows, err := r.db.Query("SELECT session_id, user_id, created_at, updated_at FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get sessions by user_id=%d: %w", userID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var s dal.Session

		if err = rows.Scan(
			&s.SessionID,
			&s.UserID,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}

		res = append(res, &s)
	}

	return res, nil
}

func (r *SessionRepository) Create(name string, userID uint64) (*dal.Session, error) {
	now := time.Now().UTC()
	res := &dal.Session{
		UserID:    userID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := r.db.QueryRow("INSERT INTO sessions (name, user_id, created_at, updated_at) VALUES ($1, $2, $3) RETURNING session_id",
		name,
		userID,
		now,
		now,
	).Scan(&res.SessionID); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return res, nil
}
