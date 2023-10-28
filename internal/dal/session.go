package dal

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dgraph-io/badger/v4"
)

const idLength = 6

var (
	idRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	errSessionAlreadyExists = errors.New("session with specified ID already exists")
)

type Session struct {
	SessionID string `json:"session_id"`
	LastUsed  int64  `json:"last_used"`
}

type SessionRepository struct {
	db *badger.DB
}

func NewSessionRepository(db *badger.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Get(key string) (*Session, error) {
	var res Session

	if err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			if err = json.Unmarshal(val, &res); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return nil
		})
	}); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &res, nil
}

func (r *SessionRepository) Create() (*Session, error) {
	s := &Session{
		SessionID: newID(),
		LastUsed:  time.Now().UnixMicro(),
	}

	if err := r.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(s.SessionID))
		if err == nil {
			return errSessionAlreadyExists
		} else if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("get: %w", err)
		}

		val, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		return txn.Set([]byte(s.SessionID), val)
	}); err != nil {
		if errors.Is(err, errSessionAlreadyExists) {
			return nil, fmt.Errorf("create session: session already exists: %w", ErrAlreadyExists)
		}

		return nil, fmt.Errorf("create session: %w", err)
	}

	return s, nil
}

func newID() string {
	b := make([]rune, idLength)

	for i := range b {
		b[i] = idRunes[rand.Intn(len(idRunes))]
	}
	return string(b)
}
