package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/log"
	"github.com/Roma7-7-7/shared-clipboard/tools"
)

const (
	passwordSaltLength = 16
)

type (
	User struct {
		ID           uint64
		Name         string
		Password     string
		PasswordSalt string
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	UserRepository interface {
		GetByName(name string) (*dal.User, error)
		Create(name, password, passwordSalt string) (*dal.User, error)
	}

	UserService struct {
		repo UserRepository
		log  log.TracedLogger
	}
)

func NewUserService(repo UserRepository, log log.TracedLogger) *UserService {
	return &UserService{
		repo: repo,
		log:  log,
	}
}

func (s *UserService) Create(ctx context.Context, name, password string) (*User, error) {
	s.log.Debugw(ctx, "creating user", "name", name)

	if err := validateSignup(name, password); err != nil {
		return nil, fmt.Errorf("validate signup: %w", err)
	}

	passwordSalt := tools.RandomAlphanumericKey(passwordSaltLength)
	salted := saltedPassword(password, passwordSalt)
	hashed, err := bcrypt.GenerateFromPassword([]byte(salted), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.Create(name, string(hashed), passwordSalt)
	if err != nil {
		if errors.Is(err, dal.ErrConflictUnique) {
			s.log.Debugw(ctx, "user with this name already exists")
			return nil, &RenderableError{
				Code:    ErrorCodeSignupConflict,
				Message: "User with specified name already exists",
			}
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	s.log.Debugw(ctx, "user created", "id", user.ID)
	return toDomainUser(user), nil
}

func (s *UserService) VerifyPassword(ctx context.Context, name, password string) (*User, error) {
	s.log.Debugw(ctx, "verifying password", "name", name)

	user, err := s.repo.GetByName(name)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			s.log.Debugw(ctx, "user not found")
			return nil, &RenderableError{
				Code:    ErrorCodeUserNotFound,
				Message: "User not found",
			}
		}

		return nil, fmt.Errorf("get user by name: %w", err)
	}

	salted := saltedPassword(password, user.PasswordSalt)
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(salted)); err != nil {
		s.log.Debugw(ctx, "wrong password")
		return nil, &RenderableError{
			Code:    ErrorCodeSiginWrongPassword,
			Message: "Wrong password",
		}
	}

	s.log.Debugw(ctx, "password verified", "id", user.ID)
	return toDomainUser(user), nil
}

func validateSignup(name, password string) *RenderableError {
	details := make(map[string]string, 2)

	if err := checkName(name); err != nil {
		details["name"] = err.Error()
	}
	if err := checkPassword(password); err != nil {
		details["password"] = err.Error()
	}

	if len(details) > 0 {
		return &RenderableError{
			Code:    ErrorCodeSignupBadRequest,
			Message: "Bad request",
			Details: details,
		}
	}

	return nil
}

func toDomainUser(dalUser *dal.User) *User {
	return &User{
		ID:           dalUser.ID,
		Name:         dalUser.Name,
		Password:     dalUser.Password,
		PasswordSalt: dalUser.PasswordSalt,
		CreatedAt:    dalUser.CreatedAt,
		UpdatedAt:    dalUser.UpdatedAt,
	}
}

func checkName(name string) error {
	if len(name) < 3 {
		return errors.New("name must be at least 3 characters long")
	}

	if len(name) > 255 {
		return errors.New("name is too long")
	}

	if !(name[0] >= 'A' && name[0] <= 'Z') && !(name[0] >= 'a' && name[0] <= 'z') {
		return errors.New("name must start with a letter")
	}

	forbiddenChars := make([]rune, 0, 10)
	for _, c := range name {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c == '_' || c == '-' || c == '.' || c == '@' || c == '+':
		default:
			forbiddenChars = append(forbiddenChars, c)
		}
	}
	if len(forbiddenChars) > 0 {
		return fmt.Errorf("name contains forbidden character(s): [%v]", string(forbiddenChars))
	}

	return nil
}

func checkPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasDigit, hasUpper, hasLower, hasSpecial bool
	)

	for _, c := range password {
		switch {
		case c >= '0' && c <= '9':
			hasDigit = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		default:
			hasSpecial = true
		}
	}

	if !hasDigit || !hasUpper || !hasLower || !hasSpecial {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, one digit and one special character")
	}

	return nil
}

func saltedPassword(password, passwordSalt string) string {
	return fmt.Sprintf("%s%s", password, passwordSalt)
}
