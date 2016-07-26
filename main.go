package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func unshorten(db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := []byte(r.URL.Path[1:][strings.LastIndex(r.URL.Path[1:], "/")+1:])
		if len(key) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("shorty"))
			if bucket == nil {
				return fmt.Errorf("Bucket shorty not found")
			}
			url := bucket.Get(key)
			if url == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil
			}
			w.Header().Add("Location", string(url))
			w.WriteHeader(http.StatusFound)
			w.Write(url)
			return nil
		})
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func shorten(host string, keycache <-chan []byte, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var key []byte
		err := db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("invshorty"))
			if err != nil {
				return err
			}
			existantKey := bucket.Get([]byte(url))
			if existantKey != nil {
				key = existantKey
				return nil
			}
			key = <-keycache
			err = bucket.Put([]byte(url), []byte(key))
			if err != nil {
				return err
			}
			bucket, err = tx.CreateBucketIfNotExists([]byte("shorty"))
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(key), []byte(url))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("http://" + host + "/s/" + string(key)))
		w.Write([]byte("\n"))
	}
}

func RandStringBytes(n int) []byte {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func keygen(keycache chan<- []byte) {
	for {
		keycache <- RandStringBytes(10)
	}
}

func main() {
	const host = "makk.es"
	keycache := make(chan []byte, 1000)
	go keygen(keycache)

	db, err := bolt.Open("shorty.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening Bolt DB: ", err)
	}
	defer db.Close()

	http.HandleFunc("/shorten", shorten(host, keycache, db))
	http.HandleFunc("/", unshorten(db))
	listener, err := net.Listen("tcp", "localhost:3002")
	if err != nil {
		log.Fatal("Error starting HTTP server", err)
	}
	log.Println("Shorty listening on " + host + ":3002")
	http.Serve(listener, nil)

}
