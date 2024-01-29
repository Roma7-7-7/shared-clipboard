package local

import (
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

const blockedJTIs = "blockedJTIs"

type JWTRepository struct {
	db *bolt.DB
}

func NewJWTRepository(db *bolt.DB) (*JWTRepository, error) {
	if err := db.Update(func(txn *bolt.Tx) error {
		return createBucket(txn, blockedJTIs)
	}); err != nil {
		return nil, err
	}

	return &JWTRepository{
		db: db,
	}, nil
}

func (r *JWTRepository) CreateBlockedJTI(jti string, expires time.Time) error {
	if err := r.db.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(blockedJTIs))

		return b.Put([]byte(jti), []byte(expires.Format(time.RFC3339)))
	}); err != nil {
		return fmt.Errorf("create blocked jti: %w", err)
	}

	return nil
}

func (r *JWTRepository) IsBlockedJTIExists(jti string) (bool, error) {
	var res bool

	if err := r.db.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(blockedJTIs))

		v := b.Get([]byte(jti))
		if v == nil {
			return nil
		}

		res = true

		return nil
	}); err != nil {
		return false, fmt.Errorf("is jti exists: %w", err)
	}

	return res, nil
}
