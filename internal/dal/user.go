package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type User struct {
	ID           uint64
	Name         string
	Password     string
	PasswordSalt string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) (*UserRepository, error) {
	return &UserRepository{
		db: db,
	}, nil
}

func (r *UserRepository) GetByID(id uint64) (*User, error) {
	var res User

	if err := r.db.QueryRow("SELECT id, name, password, password_salt, created_at, updated_at FROM users WHERE id = $1", id).Scan(
		&res.ID,
		&res.Name,
		&res.Password,
		&res.PasswordSalt,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id=%d not found: %w", id, ErrNotFound)
		}

		return nil, fmt.Errorf("get user by id=%d: %w", id, err)
	}

	return &res, nil
}

func (r *UserRepository) GetByName(name string) (*User, error) {
	var res User

	if err := r.db.QueryRow("SELECT id, name, password, password_salt, created_at, updated_at FROM users WHERE name = $1", name).Scan(
		&res.ID,
		&res.Name,
		&res.Password,
		&res.PasswordSalt,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with name=\"%s\" not found: %w", name, ErrNotFound)
		}

		return nil, fmt.Errorf("get user by name=\"%s\": %w", name, err)
	}

	return &res, nil
}

func (r *UserRepository) Create(name, password, passwordSalt string) (*User, error) {
	res := User{
		Name:         name,
		Password:     password,
		PasswordSalt: passwordSalt,
	}

	if err := r.db.QueryRow("INSERT INTO users (name, password, password_salt) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at", name, password, passwordSalt).Scan(
		&res.ID,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pgConflictErrorCode {
			return nil, fmt.Errorf("create user with name=\"%s\": %w", name, ErrConflictUnique)
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	return &res, nil

}
