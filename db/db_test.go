package db_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/makkes/shorty/db"
)

func TestErrKeyCollisionIs(t *testing.T) {
	var err error = db.NewErrKeyCollision([]byte("foobar"))

	if !errors.Is(err, db.ErrKeyCollision{}) {
		t.Fatalf("expected err to be a ErrKeyCollision")
	}

	err = fmt.Errorf("failed: %w", err)

	if !errors.Is(err, db.ErrKeyCollision{}) {
		t.Fatalf("expected err to be a ErrKeyCollision")
	}
}
