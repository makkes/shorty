package main

import (
	"testing"

	"github.com/justsocialapps/assert"
)

func TestKeygenWritesRandomKeyToChannel(t *testing.T) {
	ch := make(chan []byte)
	go keygen(ch)
	key := <-ch

	assert := assert.NewAssert(t)

	assert.Equal(len(key), 10, "Key has unexpected length")
	assert.Match("^[a-zA-Z]*$", string(key), "Key has unexpected format")
}
