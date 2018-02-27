package main

import (
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func unshorten(db DB, statch chan<- []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := []byte(r.URL.Path[1:][strings.LastIndex(r.URL.Path[1:], "/")+1:])
		if len(key) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var err error
		url, err := db.GetURL(key)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if url == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Add("Location", string(url))
		w.WriteHeader(http.StatusMovedPermanently)
		_, err = w.Write(url)
		statch <- url
	}
}

func shorten(protocol string, host string, keybuffer <-chan []byte, db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newkey []byte
		values := r.URL.Query()
		url := values.Get("url")
		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m, _ := regexp.Match("^http(s?)://", []byte(url))
		if !m {
			url = "http://" + url
		}
		_key := values.Get("key")
		if _key == "" {
			newkey = <-keybuffer
		} else {
			newkey = []byte(_key)
		}
		key, err := db.SaveURL(url, newkey)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.WriteString(w, protocol+"://"+host+"/s/"+key+"\n")
		if err != nil {
			log.Printf("Error returning shortened URL: %v", err)
		}
	}
}

func main() {
	serveHost := os.Getenv("SERVE_HOST")
	if serveHost == "" {
		serveHost = "localhost"
	}

	listenHost := os.Getenv("LISTEN_HOST")
	if listenHost == "" {
		listenHost = "localhost"
	}

	listenPort := os.Getenv("LISTEN_PORT")
	if listenPort == "" {
		listenPort = "3002"
	}

	serveProtocol := os.Getenv("SERVE_PROTOCOL")
	if serveProtocol == "" {
		serveProtocol = "https"
	}

	dbDir := os.Getenv("DB_DIR")

	rand.Seed(time.Now().UnixNano())
	keybuffer := make(chan []byte, 1000)
	go keygen(keybuffer)

	boltDb, err := bolt.Open(path.Join(dbDir, "shorty.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening Bolt DB: ", err)
	}

	defer func() {
		closeerr := boltDb.Close()
		if closeerr != nil {
			log.Printf("Error closing DB: %v", closeerr)
		}
	}()

	go stats(boltDb, func(stats string) {
		log.Println(stats)
	})

	statch := make(chan []byte)
	go collectStats(dbDir, statch)

	db := NewBoltDB(boltDb)

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/", fs)

	http.HandleFunc("/shorten", shorten(serveProtocol, serveHost, keybuffer, db))

	http.HandleFunc("/s/", unshorten(db, statch))
	listener, err := net.Listen("tcp", listenHost+":"+listenPort)
	if err != nil {
		log.Fatal("Error starting HTTP server", err)
	}
	log.Printf("Shorty listening on %s:%s\n", listenHost, listenPort)
	err = http.Serve(listener, nil)
	if err != nil {
		log.Panic(err)
	}

}
