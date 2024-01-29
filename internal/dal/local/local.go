package local

import (
	"fmt"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

func itob(v uint64) []byte {
	return []byte(strconv.FormatUint(v, 10))
}

func createBucket(txn *bolt.Tx, bucket string) error {
	_, err := txn.CreateBucketIfNotExists([]byte(bucket))
	if err != nil {
		return fmt.Errorf("create bucket %s: %w", bucket, err)
	}
	return nil
}
