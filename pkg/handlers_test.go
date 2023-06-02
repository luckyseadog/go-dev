package pkg

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
			request: "/update/gauge/b/1.0",
			want:    "<html><body><p>a</p><p>b</p></body></html>",
		},
		{
			name:    "test #3",
			request: "/update/counter/c/1",
			want:    "<html><body><p>a</p><p>b</p><p>c</p></body></html>",
		},
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(HandlerDefault))
	mux.Handle("/value/", http.HandlerFunc(HandlerGet))
	mux.Handle("/update/", http.HandlerFunc(HandlerUpdate))
	ts := httptest.NewServer(mux)
	client := ts.Client()
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, ts.URL+tt.request, nil)
			//client := ts.Client()
			res, err := client.Do(request)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			defer res.Body.Close()
			_, err = io.Copy(io.Discard, res.Body)
			assert.NoError(t, err)

			request = httptest.NewRequest(http.MethodGet, ts.URL, nil)
			res, err = client.Do(request)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, body)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.Handle("/", http.HandlerFunc(HandlerDefault))
			mux.Handle("/update/", http.HandlerFunc(HandlerUpdate))
			h := http.HandlerFunc(mux.ServeHTTP)
			h(w, request)
			result := w.Result()
			assert.Equal(t, tt.want, result.StatusCode)
			defer result.Body.Close()
			_, err := io.Copy(io.Discard, result.Body)
			assert.NoError(t, err)
		})
	}
}
