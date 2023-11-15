package dal

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

const clipboardsBucket = "clipboards"

type Clipboard struct {
	SessionID   uint64    `json:"session_id"`
	ContentType string    `json:"content_type"`
	Content     []byte    `json:"content"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ClipboardRepository struct {
	db *bolt.DB
}

func NewClipboardRepository(db *bolt.DB) (*ClipboardRepository, error) {
	if err := db.Update(func(txn *bolt.Tx) error {
		return createBucket(txn, clipboardsBucket)
	}); err != nil {
		return nil, err
	}
	return &ClipboardRepository{
		db: db,
	}, nil
}

func (r *ClipboardRepository) GetBySessionID(id uint64) (*Clipboard, error) {
	var res Clipboard

	if err := r.db.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(clipboardsBucket))

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

func (r *ClipboardRepository) SetBySessionID(id uint64, contentType string, content []byte) (*Clipboard, error) {
	var res Clipboard

	if err := r.db.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(clipboardsBucket))

		res = Clipboard{
			SessionID:   id,
			ContentType: contentType,
			Content:     content,
			UpdatedAt:   time.Now(),
		}
		v, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		return b.Put(itob(id), v)
	}); err != nil {
		return nil, err
	}

	return &res, nil
}
