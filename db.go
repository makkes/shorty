package main

import (
	"fmt"

	"github.com/boltdb/bolt"
)

// DB is the interface for bundling all database operations.
type DB interface {
	SaveURL(url string, key []byte) (string, error)
	GetURL(key []byte) ([]byte, error)
}

// A BoltDB uses Bolt to persist URLs.
type BoltDB struct {
	*bolt.DB
}

// NewBoltDB returns a BoltDB that uses db as database.
func NewBoltDB(db *bolt.DB) *BoltDB {
	return &BoltDB{db}
}

func (db *BoltDB) GetURL(key []byte) ([]byte, error) {
	var url []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("shorty"))
		if bucket == nil {
			return nil
		}
		url = bucket.Get(key)
		return nil
	})
	return url, err
}

// SaveURL saves the given url using a key from the keybuffer as short URL.
func (db *BoltDB) SaveURL(url string, key []byte) (string, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		invbucket, err := tx.CreateBucketIfNotExists([]byte("invshorty"))
		if err != nil {
			return err
		}
		existantKey := invbucket.Get([]byte(url))
		if existantKey != nil {
			return nil
		}
		bucket, err := tx.CreateBucketIfNotExists([]byte("shorty"))
		if err != nil {
			return err
		}
		if bucket.Get(key) != nil {
			return fmt.Errorf("Key collision: %s", key)
		}
		err = invbucket.Put([]byte(url), key)
		if err != nil {
			return err
		}
		err = bucket.Put(key, []byte(url))
		return err
	})
	return string(key), err
}
