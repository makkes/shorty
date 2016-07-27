package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
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
			w.WriteHeader(http.StatusMovedPermanently)
			w.Write(url)
			statch <- url
			return nil
		})
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func saveUrl(url string, keybuffer <-chan []byte, db *bolt.DB) (string, error) {
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
		err = invbucket.Put([]byte(url), []byte(key))
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(key), []byte(url))
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
		key, err := saveUrl(url, keybuffer, db)
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

func keygen(keybuffer chan<- []byte) {
	for {
		keybuffer <- RandStringBytes(10)
	}
}

func stats(db *bolt.DB, out func(string)) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		s := <-sigch
		var sum int64
		var kvpairs string = ""
		db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("shorty"))
			bucket.ForEach(func(k, v []byte) error {
				sum++
				if s == syscall.SIGUSR2 {
					kvpairs += fmt.Sprintf("\nkey=%s, value=%s", k, v)
				}
				return nil
			})
			return nil
		})
		out(fmt.Sprintf("Serving %d URLs%s", sum, kvpairs))
	}
}

func collectStats(statch <-chan []byte) {
	db, err := bolt.Open("shorty_stats.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening Bolt DB for stats: ", err)
	}
	defer db.Close()
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
	defer db.Close()

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
	http.Serve(listener, nil)

}
