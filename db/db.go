package db

import "fmt"

// DB is the interface for bundling all database operations.
type DB interface {
	SaveURL(url string, key []byte) error
	GetURL(key []byte) ([]byte, error)
}

type ErrKeyCollision struct {
	key []byte
}

func NewErrKeyCollision(key []byte) ErrKeyCollision {
	return ErrKeyCollision{
		key: key,
	}
}

func (e ErrKeyCollision) Error() string {
	return fmt.Sprintf("the key %q is already used", e.key)
}

func (e ErrKeyCollision) Is(target error) bool {
	_, ok := target.(ErrKeyCollision)
	return ok
}
