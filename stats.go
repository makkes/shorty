package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
)

func stats(db *bolt.DB, out func(string)) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		s := <-sigch
		var sum int64
		var kvpairs string = ""
		err := db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("shorty"))
			return bucket.ForEach(func(k, v []byte) error {
				sum++
				if s == syscall.SIGUSR2 {
					kvpairs += fmt.Sprintf("\nkey=%s, value=%s", k, v)
				}
				return nil
			})
		})
		if err != nil {
			log.Printf("Error collecting stats: %v", err)
		}
		out(fmt.Sprintf("Serving %d URLs%s", sum, kvpairs))
	}
}

func collectStats(statch <-chan []byte) {
	db, err := bolt.Open("shorty_stats.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening Bolt DB for stats: ", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing stats DB: %v", err)
		}

	}()
	for {
		url := <-statch
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("views"))
			if err != nil {
				return fmt.Errorf("Error opening/creating bucket 'views': %v", err)
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
			views += 1
			err = bucket.Put(url, []byte(strconv.FormatUint(views, 10)))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
}
