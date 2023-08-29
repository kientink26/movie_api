package main

import (
	"bytes"
	"io"
	"log"
	"movie_api/internal/data/mock"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
)

var router *httprouter.Router

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	os.Exit(code)
}

func setUp() {
	app := newTestApplication()
	router = httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
}

func newTestApplication() *application {
	return &application{
		logger: log.New(io.Discard, "", 0),
		models: mock.NewMockModels(),
	}
}

func requestAndGetResponse(t *testing.T, method string, urlPath string) (int, http.Header, []byte) {
	r, err := http.NewRequest(method, urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
	rs := rr.Result()

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	return rs.StatusCode, rs.Header, body
}

func TestShowMovie(t *testing.T) {
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"Valid ID", "/v1/movies/1", http.StatusOK, []byte("Black Panther")},
		{"Non-existent ID", "/v1/movies/2", http.StatusNotFound, nil},
		{"Negative ID", "/v1/movies/-1", http.StatusNotFound, nil},
		{"Decimal ID", "/v1/movies/1.23", http.StatusNotFound, nil},
		{"String ID", "/v1/movies/foo", http.StatusNotFound, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := requestAndGetResponse(t, http.MethodGet, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}
