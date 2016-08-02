package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justsocialapps/assert"
)

type TestDB struct {
	res string
	err error
	url string
}

func (db *TestDB) SaveURL(url string, keybuffer <-chan []byte) (string, error) {
	db.url = url
	return db.res, db.err
}

func setup(url string, db DB) *httptest.ResponseRecorder {
	handler := shorten("sho.rt", nil, db)
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestShortenFollowsTheHappyPath(t *testing.T) {
	w := setup("?url=THEURL", &TestDB{res: "A new key"})
	assert := assert.NewAssert(t)
	assert.Equal(w.Body.String(), "http://sho.rt/s/A new key\n", "Returned URL is incorrect")
	assert.Equal(w.Code, http.StatusOK, "Returned status code is incorrect")
}

func TestShortenDoesntAcceptEmptyURLs(t *testing.T) {
	w := setup("?url=", &TestDB{})
	assert := assert.NewAssert(t)
	assert.Equal(w.Code, http.StatusBadRequest, "Returned status code is incorrect")
}

func TestShortenShouldPrependProtocol(t *testing.T) {
	db := &TestDB{}
	setup("?url=shorty", db)

	assert := assert.NewAssert(t)
	assert.Equal(db.url, "http://shorty", "Protocol not prepended")
}

func TestShortenShouldNotPrependProtocol(t *testing.T) {
	db := &TestDB{}
	setup("?url=http://shorty", db)

	assert := assert.NewAssert(t)
	assert.Equal(db.url, "http://shorty", "URL was messed with")
}

func TestShortenHandlesDBErrorsCorrectly(t *testing.T) {
	w := setup("?url=test_url", &TestDB{res: "", err: fmt.Errorf("Error saving url")})

	assert := assert.NewAssert(t)
	assert.Equal(w.Code, http.StatusInternalServerError, "Returned status code is incorrect")
}
