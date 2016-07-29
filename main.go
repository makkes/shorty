package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func unshorten(db *bolt.DB, statch chan<- []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := []byte(r.URL.Path[1:][strings.LastIndex(r.URL.Path[1:], "/")+1:])
		if len(key) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var err error
		err = db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("shorty"))
			if bucket == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil
			}
			url := bucket.Get(key)
			if url == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil
			}
			w.Header().Add("Location", string(url))
			w.WriteHeader(http.StatusMovedPermanently)
			_, err = w.Write(url)
			statch <- url
			return err
		})
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func saveURL(url string, keybuffer <-chan []byte, db *bolt.DB) (string, error) {
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
		if err != nil {
			return err
		}
		return nil
	})
	return string(key), err
}

func shorten(host string, keybuffer <-chan []byte, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m, _ := regexp.Match("^http(s?)://", []byte(url))
		if !m {
			url = "http://" + url
		}
		key, err := saveURL(url, keybuffer, db)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.WriteString(w, "http://"+host+"/s/"+key+"\n")
		if err != nil {
			log.Printf("Error returning shortened URL: %v", err)
		}
	}
}

func main() {
	host := flag.String("host", "localhost", "The hostname used to reach Shorty")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())
	keybuffer := make(chan []byte, 1000)
	go keygen(keybuffer)

	db, err := bolt.Open("shorty.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening Bolt DB: ", err)
	}
	defer func() {
		closeerr := db.Close()
		if closeerr != nil {
			log.Printf("Error closing DB: %v", closeerr)
		}

	}()

	go stats(db, func(stats string) {
		log.Println(stats)
	})

	statch := make(chan []byte)
	go collectStats(statch)

	http.HandleFunc("/shorten", shorten(*host, keybuffer, db))
	http.HandleFunc("/", unshorten(db, statch))
	listener, err := net.Listen("tcp", "localhost:3002")
	if err != nil {
		log.Fatal("Error starting HTTP server", err)
	}
	log.Println("Shorty listening on " + *host + ":3002")
	err = http.Serve(listener, nil)
	if err != nil {
		log.Panic(err)
	}

}
