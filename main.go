package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/makkes/shorty/boltdb"
	"github.com/makkes/shorty/db"
	dbpkg "github.com/makkes/shorty/db"
)

func unshorten(db dbpkg.DB) http.HandlerFunc {
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
			log.Printf("no URL found for key %q", key)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Add("Location", string(url))
		w.WriteHeader(http.StatusMovedPermanently)
		_, err = w.Write(url)
		if err != nil {
			log.Printf("failed writing response: %v", err)
		}
	}
}

func info(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, "This is Shorty, running on "+hostname+"\n")
}

func shorten(protocol string, host string, keybuffer <-chan []byte, db dbpkg.DB) http.HandlerFunc {
	urlProtoRE := regexp.MustCompile("^http(s?)://")

	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !urlProtoRE.Match([]byte(url)) {
			url = "http://" + url
		}

		key := []byte(r.URL.Query().Get("key"))
		if len(key) == 0 {
			key = <-keybuffer
		}

		err := db.SaveURL(url, key)
		if err != nil {
			if errors.Is(err, dbpkg.ErrKeyCollision{}) {
				http.Error(w, fmt.Sprintf("key %q is already used", key), http.StatusConflict)
				return
			}
			log.Printf("failed saving URL to DB: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = fmt.Fprintf(w, "%s://%s/s/%s\n", protocol, host, key)
		if err != nil {
			log.Printf("Error returning shortened URL: %v", err)
		}
	}
}

func main() {

	backends := map[string]func() (db.DB, error){
		"bolt": boltdb.NewBoltDB,
	}

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

	backend := os.Getenv("BACKEND")
	if backend == "" {
		backend = "bolt"
	}

	keybuffer := make(chan []byte, 1000)
	go keygen(keybuffer)

	db, err := backends[backend]()
	if err != nil {
		log.Fatalf("Error creating DB backend: %s", err)
	}

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/", fs)

	http.HandleFunc("/shorten", shorten(serveProtocol, serveHost, keybuffer, db))
	http.HandleFunc("/info", info)

	http.HandleFunc("/s/", unshorten(db))
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
