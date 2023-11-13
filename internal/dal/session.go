package dal

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
)

const idLength = 6

var (
	idRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type contentType string

type Session struct {
	SessionID   uint64    `json:"session_id"`
	JoinKey     string    `json:"join_key"`
	ContentType string    `json:"content_type"`
	Content     []byte    `json:"content"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SessionRepository struct {
	db           *badger.DB
	sessionIDSeq *badger.Sequence
}

func NewSessionRepository(db *badger.DB) (*SessionRepository, error) {
	sessionIDSequence, err := db.GetSequence([]byte("session"), 100)
	if err != nil {
		return nil, fmt.Errorf("get sequence: %w", err)
	}
	return &SessionRepository{
		db:           db,
		sessionIDSeq: sessionIDSequence,
	}, nil
}

func (r *SessionRepository) GetByID(id uint64) (*Session, error) {
	var res Session

	if err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(sessionDBKey(id))
		if err != nil {
			return fmt.Errorf("get session: %w", err)
		}

		if err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &res)
		}); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		return nil
	}); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &res, nil
}

func (r *SessionRepository) GetByJoinKey(key string) (*Session, error) {
	var res Session

	if err := r.db.View(func(txn *badger.Txn) error {
		mappingItem, err := txn.Get(joinDBKey(key))
		if err != nil {
			return fmt.Errorf("get join key: %w", err)
		}

		var sDBKey []byte
		if err = mappingItem.Value(func(val []byte) error {
			sDBKey = val
			return nil
		}); err != nil {
			return fmt.Errorf("get session key: %w", err)
		}

		sessionItem, err := txn.Get(sDBKey)
		if err != nil {
			return fmt.Errorf("get session: %w", err)
		}

		if err = sessionItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &res)
		}); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		return err
	}); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &res, nil
}

func (r *SessionRepository) Create() (*Session, error) {
	sid, err := r.sessionIDSeq.Next()
	if err != nil {
		return nil, fmt.Errorf("next session ID: %w", err)
	}
	s := &Session{
		SessionID: sid,
		JoinKey:   randomAlphanumericKey(),
		UpdatedAt: time.Now().UTC(),
	}

	if err = r.db.Update(func(txn *badger.Txn) error {
		val, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		sDBKey := sessionDBKey(s.SessionID)
		if err = txn.Set(sDBKey, val); err != nil {
			return fmt.Errorf("set session: %w", err)
		}

		jDBKey := joinDBKey(s.JoinKey)
		if _, err = txn.Get(jDBKey); err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return fmt.Errorf("get join key: %w", err)
			}
		} else {
			return errors.New("join key already exists")
		}

		if err = txn.Set(jDBKey, sDBKey); err != nil {
			return fmt.Errorf("set join key: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return s, nil
}

func (r *SessionRepository) GetContentByID(sid uint64) (*Session, error) {
	var s Session

	if err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(sessionDBKey(sid))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotFound
			}

			return fmt.Errorf("get session: %w", err)
		}

		if err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &s)
		}); err != nil {
			return fmt.Errorf("get session: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	return &s, nil
}

func (r *SessionRepository) SetContentByID(sid uint64, contentType string, content []byte) (*Session, error) {
	var s Session

	if err := r.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(sessionDBKey(sid))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotFound
			}

			return fmt.Errorf("get session: %w", err)
		}

		if err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &s)
		}); err != nil {
			return fmt.Errorf("get session: %w", err)
		}

		s.ContentType = contentType
		s.Content = content
		s.UpdatedAt = time.Now().UTC()

		val, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		return txn.Set(sessionDBKey(s.SessionID), val)
	}); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	return &s, nil
}

func (r *SessionRepository) mapJoinKeyToSessionID(key string, sessionID uint64) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set(joinDBKey(key), []byte(strconv.FormatUint(sessionID, 10)))
	})
}

func sessionDBKey(id uint64) []byte {
	return []byte(fmt.Sprintf("sessions#%s", strconv.FormatUint(id, 10)))
}

func joinDBKey(key string) []byte {
	return []byte(fmt.Sprintf("joins#%s", key))
}

func randomAlphanumericKey() string {
	b := make([]rune, idLength)

	for i := range b {
		b[i] = idRunes[rand.Intn(len(idRunes))]
	}
	return string(b)
}
