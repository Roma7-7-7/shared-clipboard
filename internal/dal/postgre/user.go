package postgre

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) (*UserRepository, error) {
	return &UserRepository{
		db: db,
	}, nil
}

func (r *UserRepository) GetByID(id uint64) (*dal.User, error) {
	var res dal.User

	if err := r.db.QueryRow("SELECT user_id, name, password, password_salt, created_at, updated_at FROM users WHERE user_id = $1", id).Scan(
		&res.ID,
		&res.Name,
		&res.Password,
		&res.PasswordSalt,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id=%d not found: %w", id, dal.ErrNotFound)
		}

		return nil, fmt.Errorf("get user by user_id=%d: %w", id, err)
	}

	return &res, nil
}

func (r *UserRepository) GetByName(name string) (*dal.User, error) {
	var res dal.User

	if err := r.db.QueryRow("SELECT user_id, name, password, password_salt, created_at, updated_at FROM users WHERE name = $1", name).Scan(
		&res.ID,
		&res.Name,
		&res.Password,
		&res.PasswordSalt,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with name=\"%s\" not found: %w", name, dal.ErrNotFound)
		}

		return nil, fmt.Errorf("get user by name=\"%s\": %w", name, err)
	}

	return &res, nil
}

func (r *UserRepository) Create(name, password, passwordSalt string) (*dal.User, error) {
	res := dal.User{
		Name:         name,
		Password:     password,
		PasswordSalt: passwordSalt,
	}

	if err := r.db.QueryRow("INSERT INTO users (name, password, password_salt) VALUES ($1, $2, $3) RETURNING user_id, created_at, updated_at", name, password, passwordSalt).Scan(
		&res.ID,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pgConflictErrorCode {
			return nil, fmt.Errorf("create user with name=\"%s\": %w", name, dal.ErrConflictUnique)
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	return &res, nil

}
