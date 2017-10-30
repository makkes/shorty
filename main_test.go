package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justsocialapps/assert"
)

type TestDB struct {
	res     string
	err     error
	url     string
	longURL []byte
}

func (db *TestDB) SaveURL(url string, keybuffer <-chan []byte) (string, error) {
	db.url = url
	return db.res, db.err
}

func (db *TestDB) GetURL(key []byte) ([]byte, error) {
	return db.longURL, db.err
}

func setupShorten(url, proto string, db DB) *httptest.ResponseRecorder {
	handler := shorten(proto, "sho.rt", nil, db)
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func setupUnshorten(url string, db DB, statch chan<- []byte) *httptest.ResponseRecorder {
	handler := unshorten(db, statch)
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestShortenFollowsTheHappyPath(t *testing.T) {
	w := setupShorten("?url=THEURL", "http", &TestDB{res: "A new key"})
	assert := assert.NewAssert(t)
	assert.Equal(w.Body.String(), "http://sho.rt/s/A new key\n", "Returned URL is incorrect")
	assert.Equal(w.Code, http.StatusOK, "Returned status code is incorrect")
}

func TestShortenDoesntAcceptEmptyURLs(t *testing.T) {
	w := setupShorten("?url=", "http", &TestDB{})
	assert := assert.NewAssert(t)
	assert.Equal(w.Code, http.StatusBadRequest, "Returned status code is incorrect")
}

func TestShortenRespectsTheProtocol(t *testing.T) {
	w := setupShorten("?url=THEURL", "http", &TestDB{res: "A new key"})
	a := assert.NewAssert(t)
	a.Equal(w.Body.String(), "http://sho.rt/s/A new key\n", "Returned URL is incorrect")
	a.Equal(w.Code, http.StatusOK, "Returned status code is incorrect")

	w = setupShorten("?url=THEURL", "https", &TestDB{res: "A new key"})
	a = assert.NewAssert(t)
	a.Equal(w.Body.String(), "https://sho.rt/s/A new key\n", "Returned URL is incorrect")
	a.Equal(w.Code, http.StatusOK, "Returned status code is incorrect")
}

func TestShortenShouldPrependProtocol(t *testing.T) {
	db := &TestDB{}
	setupShorten("?url=shorty", "http", db)

	assert := assert.NewAssert(t)
	assert.Equal(db.url, "http://shorty", "Protocol not prepended")
}

func TestShortenShouldNotPrependProtocol(t *testing.T) {
	db := &TestDB{}
	setupShorten("?url=http://shorty", "http", db)

	assert := assert.NewAssert(t)
	assert.Equal(db.url, "http://shorty", "URL was messed with")
}

func TestShortenHandlesDBErrorsCorrectly(t *testing.T) {
	w := setupShorten("?url=test_url", "http", &TestDB{res: "", err: fmt.Errorf("Error saving url")})

	assert := assert.NewAssert(t)
	assert.Equal(w.Code, http.StatusInternalServerError, "Returned status code is incorrect")
}

func TestUnshortenFollowsTheHappyPath(t *testing.T) {
	statch := make(chan []byte, 1)
	w := setupUnshorten("/unshorten/veryShort", &TestDB{longURL: []byte("TheLongURL")}, statch)
	assert := assert.NewAssert(t)

	assert.Equal(w.Header().Get("Location"), "TheLongURL", "Returned long URL is incorrect")
	assert.Equal(string(<-statch), "TheLongURL", "unshorten wrote wrong URL to the stat channel")
}

func TestUnshortenHandlesUnknownKeysCorrectly(t *testing.T) {
	w := setupUnshorten("/unshorten/veryShort", &TestDB{longURL: nil}, nil)
	assert := assert.NewAssert(t)

	assert.Equal(w.Code, http.StatusNotFound, "Returned HTTP status is incorrect")
}

func TestUnshortenHandlesWrongKeysCorrectly(t *testing.T) {
	w := setupUnshorten("/no/key/", &TestDB{longURL: nil}, nil)
	assert := assert.NewAssert(t)

	assert.Equal(w.Code, http.StatusBadRequest, "Returned HTTP status is incorrect")
}

func TestUnshortenHandlesDBErrorsCorrectly(t *testing.T) {
	w := setupUnshorten("/key", &TestDB{err: fmt.Errorf("Error retrieving long URL")}, nil)
	assert := assert.NewAssert(t)

	assert.Equal(w.Code, http.StatusInternalServerError, "Returned HTTP status is incorrect")
}
