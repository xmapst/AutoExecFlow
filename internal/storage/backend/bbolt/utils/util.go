package utils

import (
	"bytes"

	"go.etcd.io/bbolt"
)

func Join(b ...[]byte) []byte {
	return bytes.Join(b, []byte("#"))
}

func Int64ToBytes(id int64) []byte {
	return []byte{
		byte(id >> 56),
		byte(id >> 48),
		byte(id >> 40),
		byte(id >> 32),
		byte(id >> 24),
		byte(id >> 16),
		byte(id >> 8),
		byte(id),
	}
}

func Bucket(tx *bbolt.Tx, prefixes ...[]byte) (*bbolt.Bucket, error) {
	if tx == nil || len(prefixes) == 0 {
		return nil, bbolt.ErrBucketNotFound
	}
	bucket := tx.Bucket(prefixes[0])
	if bucket == nil {
		return nil, bbolt.ErrBucketNotFound
	}
	if len(prefixes) == 1 {
		return bucket, nil
	}
	for _, prefix := range prefixes[1:] {
		bucket = bucket.Bucket(prefix)
		if bucket == nil {
			return nil, bbolt.ErrBucketNotFound
		}
	}
	return bucket, nil
}

func CreateBucketIfNotExists(tx *bbolt.Tx, prefixes ...[]byte) (*bbolt.Bucket, error) {
	if tx == nil || len(prefixes) == 0 {
		return nil, bbolt.ErrBucketNotFound
	}
	bucket, err := tx.CreateBucketIfNotExists(prefixes[0])
	if err != nil {
		return nil, err
	}
	if len(prefixes) == 1 {
		return bucket, nil
	}
	for _, prefix := range prefixes[1:] {
		bucket, err = bucket.CreateBucketIfNotExists(prefix)
		if err != nil {
			return nil, err
		}
	}
	return bucket, nil
}
