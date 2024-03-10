package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"
)

type (
	SessionFilter struct {
		userID          uint64
		limit           int
		name            string
		sortBy          string
		sortByDirection string
		offset          int
	}

	Session struct {
		ID        uint64
		Name      string
		UserID    uint64
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	SessionRepository struct {
		db *sql.DB
	}
)

func (s SessionFilter) UserID() uint64 {
	return s.userID
}

func (s SessionFilter) Limit() int {
	return s.limit
}

func (s SessionFilter) Name() string {
	return s.name
}

func (s SessionFilter) SortBy() string {
	return s.sortBy
}

func (s SessionFilter) SortByDirection() string {
	return s.sortByDirection
}

func (s SessionFilter) Offset() int {
	return s.offset
}

func WithName(name string) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.name = name
		return f
	}
}

func WithSortByName(desc bool) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.sortBy = "name"
		if desc {
			f.sortByDirection = "DESC"
		} else {
			f.sortByDirection = "ASC"
		}
		return f
	}
}

func WithSortByUpdateAt(desc bool) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.sortBy = "updated_at"
		if desc {
			f.sortByDirection = "DESC"
		} else {
			f.sortByDirection = "ASC"
		}
		return f
	}
}

func WithOffset(offset int) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.offset = offset
		return f
	}
}

func NewSessionFilter(userID uint64, limit int, opts ...func(SessionFilter) SessionFilter) SessionFilter {
	f := SessionFilter{
		userID:          userID,
		limit:           limit,
		name:            "",
		sortBy:          "updated_at",
		sortByDirection: "ASC",
		offset:          0,
	}

	for _, opt := range opts {
		f = opt(f)
	}

	return f
}

func NewSessionRepository(db *sql.DB) (*SessionRepository, error) {
	return &SessionRepository{
		db: db,
	}, nil
}

func (r *SessionRepository) GetByID(id uint64) (*Session, error) {
	var res Session

	if err := r.db.QueryRow("SELECT session_id, user_id, name, created_at, updated_at FROM sessions WHERE session_id = $1", id).
		Scan(
			&res.ID,
			&res.UserID,
			&res.Name,
			&res.CreatedAt,
			&res.UpdatedAt,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session with session_id=%d not found: %w", id, ErrNotFound)
		}

		return nil, fmt.Errorf("get session by session_id=%d: %w", id, err)
	}

	return &res, nil
}

func (r *SessionRepository) GetAllByUserID(userID uint64) ([]*Session, error) {
	res := make([]*Session, 0, 10)

	rows, err := r.db.Query("SELECT session_id, user_id, name, created_at, updated_at FROM sessions WHERE user_id = $1 ORDER BY updated_at DESC", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get sessions by user_id=%d: %w", userID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var s Session

		if err = rows.Scan(
			&s.ID,
			&s.UserID,
			&s.Name,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}

		res = append(res, &s)
	}

	return res, nil
}

func (r *SessionRepository) FilterBy(filter SessionFilter) ([]*Session, int, error) {
	var (
		totalCount      = 0
		res             = make([]*Session, 0, 10)
		filterQuery     = "user_id = $1 AND ($2 = '' OR name LIKE $2)"
		totalCountQuery = "SELECT COUNT(*) FROM sessions WHERE " + filterQuery
		query           = fmt.Sprintf("SELECT session_id, user_id, name, created_at, updated_at FROM sessions WHERE "+filterQuery+" ORDER BY %s %s OFFSET $3 LIMIT $4", filter.SortBy(), filter.SortByDirection())
		totalCountErr   error
		queryErr        error
	)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := r.db.QueryRow(totalCountQuery, filter.UserID(), "%"+filter.Name()+"%").Scan(&totalCount); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				totalCount = 0
				return
			}

			totalCountErr = fmt.Errorf("get total count: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		rows, err := r.db.Query(query, filter.UserID(), "%"+filter.Name()+"%", filter.Offset(), filter.Limit())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return
			}

			queryErr = fmt.Errorf("get sessions: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var s Session

			if err = rows.Scan(
				&s.ID,
				&s.UserID,
				&s.Name,
				&s.CreatedAt,
				&s.UpdatedAt,
			); err != nil {
				queryErr = fmt.Errorf("scan session: %w", err)
				return
			}

			res = append(res, &s)
		}
	}()

	wg.Wait()
	if totalCountErr != nil && queryErr != nil {
		return nil, 0, fmt.Errorf("filter sessions: %w; %w", totalCountErr, queryErr)
	}
	if totalCountErr != nil {
		return nil, 0, fmt.Errorf("filter sessions: %w", totalCountErr)
	}
	if queryErr != nil {
		return nil, 0, fmt.Errorf("filter sessions: %w", queryErr)
	}

	return res, totalCount, nil
}

func (r *SessionRepository) Create(name string, userID uint64) (*Session, error) {
	res := &Session{
		UserID: userID,
		Name:   name,
	}

	if err := r.db.QueryRow("INSERT INTO sessions (name, user_id, created_at, updated_at) VALUES ($1, $2, now(), now()) RETURNING session_id, created_at, updated_at",
		name,
		userID,
	).Scan(
		&res.ID,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return res, nil
}

func (r *SessionRepository) Update(id uint64, name string) (*Session, error) {
	execRes, err := r.db.Exec("UPDATE sessions SET name = $1, updated_at = now() WHERE session_id = $2",
		name,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session with session_id=%d not found: %w", id, ErrNotFound)
		}

		return nil, fmt.Errorf("update session: %w", err)
	}

	affected, err := execRes.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("session with session_id=%d not found: %w", id, ErrNotFound)
	}

	return r.GetByID(id)
}

func (r *SessionRepository) UpdateUpdatedAt(id uint64) error {
	execRes, err := r.db.Exec("UPDATE sessions SET updated_at = now() WHERE session_id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("session with session_id=%d not found: %w", id, ErrNotFound)
		}

		return fmt.Errorf("update session: %w", err)
	}

	affected, err := execRes.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("session with session_id=%d not found: %w", id, ErrNotFound)
	}

	return nil
}

func (r *SessionRepository) Delete(id uint64) error {
	execRes, err := r.db.Exec("DELETE FROM sessions WHERE session_id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("session with session_id=%d not found: %w", id, ErrNotFound)
		}

		return fmt.Errorf("delete session: %w", err)
	}

	affected, err := execRes.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("session with session_id=%d not found: %w", id, ErrNotFound)
	}

	return nil
}
