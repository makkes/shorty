package main

// DB is the interface for bundling all database operations.
type DB interface {
	SaveURL(url string, keybuffer <-chan []byte) (string, error)
	GetURL(key []byte) ([]byte, error)
}
