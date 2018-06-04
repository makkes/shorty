package boltdb

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/makkes/shorty/db"
)

// A BoltDB uses Bolt to persist URLs.
type BoltDB struct {
	*bolt.DB
}

// NewBoltDB returns a BoltDB that uses db as database.
func NewBoltDB() (db.DB, error) {
	dbDir := os.Getenv("DB_DIR")
	db, err := bolt.Open(path.Join(dbDir, "shorty.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error opening Bolt DB: %s", err))
	}

	return &BoltDB{db}, nil
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
func (db *BoltDB) SaveURL(url string, keybuffer <-chan []byte) (string, error) {
	var key []byte
	err := db.Update(func(tx *bolt.Tx) error {
		invbucket, err := tx.CreateBucketIfNotExists([]byte("invshorty"))
		if err != nil {
			return err
		}
		existantKey := invbucket.Get([]byte(url))
		if existantKey != nil {
			key = existantKey
			return nil
		}
		key = <-keybuffer
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

func collectStats(dbDir string, statch <-chan []byte) {
	db, err := bolt.Open(path.Join(dbDir, "shorty_stats.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening Bolt DB for stats: ", err)
	}
	defer func() {
		closeerr := db.Close()
		if err != nil {
			log.Printf("Error closing stats DB: %v", closeerr)
		}

	}()
	for {
		url := <-statch
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, berr := tx.CreateBucketIfNotExists([]byte("views"))
			if berr != nil {
				return fmt.Errorf("Error opening/creating bucket 'views': %v", berr)
			}
			viewBytes := bucket.Get(url)
			var views uint64
			if viewBytes != nil {
				views, err = strconv.ParseUint(string(viewBytes), 10, 64)
				if err != nil {
					return fmt.Errorf("Error decoding views for %s: %v", string(url), err)
				}
			} else {
				views = 0
			}
			views++
			err = bucket.Put(url, []byte(strconv.FormatUint(views, 10)))
			return err
		})
		if err != nil {
			log.Println(err)
		}
	}
}
