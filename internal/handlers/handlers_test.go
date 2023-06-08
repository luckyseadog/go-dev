package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandlerDefault(t *testing.T) {
	tests := []struct {
		name    string
		request string
		want    string
	}{
		{
			name:    "test #1",
			request: "/update/gauge/a/1.0",
			want:    "<html><body><p>a</p></body></html>",
		},
		{
			name:    "test #2",
			request: "/update/counter/c/1",
			want:    "<html><body><p>a</p><p>c</p></body></html>",
		},
	}
	r := chi.NewRouter()
	r.Get("/", HandlerDefault)
	r.Get("/value/*", HandlerGet)
	r.Post("/update/*", HandlerUpdate)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := http.Post(ts.URL+tt.request, "text/plain", nil)
			assert.NoError(t, err)
			defer response.Body.Close()

			assert.Equal(t, http.StatusOK, response.StatusCode)
			_, err = io.Copy(io.Discard, response.Body)
			assert.NoError(t, err)

			log.Println(ts.URL)
			response, err = http.Get(ts.URL)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(body))
		})
	}
}

func TestHandlerUpdate(t *testing.T) {
	tests := []struct {
		name    string
		request string
		want    int
	}{
		{
			name:    "simple test #1",
			want:    http.StatusOK,
			request: "http://127.0.0.1:8080/update/gauge/Alloc/1.0",
		},
		{
			name:    "simple test #2",
			want:    http.StatusNotImplemented,
			request: "http://127.0.0.1:8080/update/unknown/Alloc/1.0",
		},
		{
			name:    "simple test #3",
			want:    http.StatusBadRequest,
			request: "http://127.0.0.1:8080/update/gauge/Alloc/hello",
		},
		{
			name:    "simple test #4",
			want:    http.StatusOK,
			request: "http://127.0.0.1:8080/update/counter/testCounter/100",
		},
		{
			name:    "simple test #4",
			want:    http.StatusBadRequest,
			request: "http://127.0.0.1:8080/update/counter/testCounter/none",
		},
		{
			name:    "simple test #5",
			want:    http.StatusNotFound,
			request: "http://127.0.0.1:8080/updater/counter/testCounter/1",
		},
		{
			name:    "simple test #6",
			want:    http.StatusNotFound,
			request: "http://127.0.0.1:8080/update/gauge/",
		},
	}
	r := chi.NewRouter()
	r.Get("/", HandlerDefault)
	r.Get("/value/*", HandlerGet)
	r.Post("/update/*", HandlerUpdate)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(r.ServeHTTP)
			h(w, request)
			result := w.Result()
			assert.Equal(t, tt.want, result.StatusCode)
			defer result.Body.Close()
			_, err := io.Copy(io.Discard, result.Body)
			assert.NoError(t, err)
		})
	}
}
