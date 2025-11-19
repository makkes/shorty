package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/makkes/shorty/assert"
	"github.com/makkes/shorty/db"
)

type TestDB struct {
	saveErr error
	getErr  error
	key     []byte
	url     []byte
}

func (tdb *TestDB) SaveURL(url string, key []byte) error {
	tdb.key = key
	tdb.url = []byte(url)
	return tdb.saveErr
}

func (tdb *TestDB) GetURL(key []byte) ([]byte, error) {
	if tdb.getErr != nil {
		return nil, tdb.getErr
	}

	if slices.Equal(tdb.key, key) {
		return tdb.url, nil
	}
	return nil, nil
}

func (tdb *TestDB) GetStats() (db.Stats, error) {
	return db.Stats{}, nil
}

func setupShorten(url, key, proto string, db db.DB) *httptest.ResponseRecorder {
	keybuffer := make(chan []byte, 1)
	keybuffer <- []byte(key)
	handler := shorten(proto, "sho.rt", keybuffer, db)
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func setupUnshorten(url string, db db.DB) *httptest.ResponseRecorder {
	handler := unshorten(db)
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestInfoReturnsInfoAboutTheRunningInstance(t *testing.T) {
	req, _ := http.NewRequest("GET", "/info", nil)
	w := httptest.NewRecorder()
	http.HandlerFunc(info(&TestDB{})).ServeHTTP(w, req)

	assert := assert.NewAssert(t)
	assert.Equal(w.Code, http.StatusOK, "Unexpected HTTP status")
	assert.Match("This is Shorty ", w.Body.String(), "Unexpected body")
}

func TestShortenFollowsTheHappyPath(t *testing.T) {
	w := setupShorten("?url=THEURL", "A new key", "http", &TestDB{})
	assert := assert.NewAssert(t)
	assert.Equal(w.Body.String(), "http://sho.rt/A new key\n", "Returned URL is incorrect")
	assert.Equal(w.Code, http.StatusOK, "Returned status code is incorrect")
}

func TestShortenDoesntAcceptEmptyURLs(t *testing.T) {
	w := setupShorten("?url=", "", "http", &TestDB{})
	assert := assert.NewAssert(t)
	assert.Equal(w.Code, http.StatusBadRequest, "Returned status code is incorrect")
}

func TestShortenRespectsTheProtocol(t *testing.T) {
	w := setupShorten("?url=THEURL", "A new key", "http", &TestDB{})
	a := assert.NewAssert(t)
	a.Equal(w.Body.String(), "http://sho.rt/A new key\n", "Returned URL is incorrect")
	a.Equal(w.Code, http.StatusOK, "Returned status code is incorrect")

	w = setupShorten("?url=THEURL", "A new key", "https", &TestDB{})
	a = assert.NewAssert(t)
	a.Equal(w.Body.String(), "https://sho.rt/A new key\n", "Returned URL is incorrect")
	a.Equal(w.Code, http.StatusOK, "Returned status code is incorrect")
}

func TestShortenShouldPrependProtocol(t *testing.T) {
	db := &TestDB{}
	setupShorten("?url=shorty", "", "http", db)

	assert := assert.NewAssert(t)
	assert.Equal(string(db.url), "http://shorty", "Protocol not prepended")
}

func TestShortenShouldNotPrependProtocol(t *testing.T) {
	db := &TestDB{}
	setupShorten("?url=http://shorty", "", "http", db)

	assert := assert.NewAssert(t)
	assert.Equal(string(db.url), "http://shorty", "URL was messed with")
}

func TestShortenHandlesDBErrorsCorrectly(t *testing.T) {
	w := setupShorten("?url=test_url", "", "http", &TestDB{saveErr: fmt.Errorf("Error saving url")})

	assert := assert.NewAssert(t)
	assert.Equal(w.Code, http.StatusInternalServerError, "Returned status code is incorrect")
}

func TestShortenCorrectlyStoresProvidedKey(t *testing.T) {
	w := setupShorten("?url=test_url&key=my_key", "", "http", &TestDB{})
	a := assert.NewAssert(t)
	a.Equal(w.Code, http.StatusOK, "returned status code is incorrect")
	a.Equal(w.Body.String(), "http://sho.rt/my_key\n", "returned URL is incorrect")
}

func TestUnshortenFollowsTheHappyPath(t *testing.T) {
	w := setupUnshorten("/unshorten/veryShort", &TestDB{
		key: []byte("veryShort"),
		url: []byte("TheLongURL"),
	})
	assert := assert.NewAssert(t)

	assert.Equal(w.Header().Get("Location"), "TheLongURL", "Returned long URL is incorrect")
}

func TestUnshortenHandlesUnknownKeysCorrectly(t *testing.T) {
	w := setupUnshorten("/unshorten/veryShort", &TestDB{})
	assert := assert.NewAssert(t)

	assert.Equal(w.Code, http.StatusNotFound, "Returned HTTP status is incorrect")
}

func TestUnshortenHandlesWrongKeysCorrectly(t *testing.T) {
	w := setupUnshorten("/no/key/", &TestDB{})
	assert := assert.NewAssert(t)

	assert.Equal(w.Code, http.StatusBadRequest, "Returned HTTP status is incorrect")
}

func TestUnshortenHandlesDBErrorsCorrectly(t *testing.T) {
	w := setupUnshorten("/key", &TestDB{getErr: fmt.Errorf("Error retrieving long URL")})
	assert := assert.NewAssert(t)

	assert.Equal(w.Code, http.StatusInternalServerError, "Returned HTTP status is incorrect")
}
