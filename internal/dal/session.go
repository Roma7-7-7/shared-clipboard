package dal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	idLength       = 6
	sessionsBucket = "sessions"
	joinKeysBucket = "joinKeys"
)

var (
	joinKeyRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type Session struct {
	SessionID uint64    `json:"session_id"`
	JoinKey   string    `json:"join_key"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SessionRepository struct {
	db *bolt.DB
}

func NewSessionRepository(db *bolt.DB) (*SessionRepository, error) {
	if err := db.Update(func(txn *bolt.Tx) error {
		for _, bucket := range []string{sessionsBucket, joinKeysBucket} {
			if err := createBucket(txn, bucket); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &SessionRepository{
		db: db,
	}, nil
}

func (r *SessionRepository) GetByID(id uint64) (*Session, error) {
	var res Session

	if err := r.db.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(sessionsBucket))

		v := b.Get(itob(id))
		if v == nil {
			return ErrNotFound
		}

		if err := json.Unmarshal(v, &res); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *SessionRepository) GetByJoinKey(key string) (*Session, error) {
	var res Session

	if err := r.db.View(func(txn *bolt.Tx) error {
		sid := txn.Bucket([]byte(joinKeysBucket)).Get([]byte(key))
		if sid == nil {
			return ErrNotFound
		}

		b := txn.Bucket([]byte(sessionsBucket))
		sb := b.Get(sid)
		if sb == nil {
			return ErrNotFound
		}

		if err := json.Unmarshal(sb, &res); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *SessionRepository) Create() (*Session, error) {
	var res *Session

	if err := r.db.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(sessionsBucket))
		sid, err := b.NextSequence()
		if err != nil {
			return fmt.Errorf("next sequence: %w", err)
		}

		res = &Session{
			SessionID: sid,
			JoinKey:   randomAlphanumericKey(),
			UpdatedAt: time.Now().UTC(),
		}

		bytes, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		if err = b.Put(itob(sid), bytes); err != nil {
			return fmt.Errorf("put session: %w", err)
		}

		b = txn.Bucket([]byte(joinKeysBucket))
		if err = b.Put([]byte(res.JoinKey), itob(sid)); err != nil {
			return fmt.Errorf("put join key: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return res, nil
}

func randomAlphanumericKey() string {
	b := make([]rune, idLength)

	for i := range b {
		b[i] = joinKeyRunes[rand.Intn(len(joinKeyRunes))]
	}
	return string(b)
}
