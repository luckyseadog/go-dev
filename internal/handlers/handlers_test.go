package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
	"github.com/stretchr/testify/require"

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
	s := storage.NewStorage()
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		HandlerDefault(w, r, s)
	})
	r.Get("/value/{^+}/*", func(w http.ResponseWriter, r *http.Request) {
		HandlerGet(w, r, s)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			HandlerValueJSON(w, r, s, []byte{})
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			HandlerValueJSON(w, r, s, []byte{})
		})
	})

	r.Post("/update/{^+}/*", func(w http.ResponseWriter, r *http.Request) {
		HandlerUpdate(w, r, s)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			HandlerUpdateJSON(w, r, s, []byte{})
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			HandlerUpdateJSON(w, r, s, []byte{})
		})
	})

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
	s := storage.NewStorage()
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		HandlerDefault(w, r, s)
	})
	r.Get("/value/{^+}/*", func(w http.ResponseWriter, r *http.Request) {
		HandlerGet(w, r, s)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			HandlerValueJSON(w, r, s, []byte{})
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			HandlerValueJSON(w, r, s, []byte{})
		})
	})

	r.Post("/update/{^+}/*", func(w http.ResponseWriter, r *http.Request) {
		HandlerUpdate(w, r, s)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			HandlerUpdateJSON(w, r, s, []byte{})
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			HandlerUpdateJSON(w, r, s, []byte{})
		})
	})

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

func TestHandlerUpdateJSON(t *testing.T) {
	tests := []struct {
		name          string
		request       string
		want          int
		bodies        [][]byte
		answerGauge   []metrics.Gauge
		answerCounter []metrics.Counter
	}{
		{
			name:    "test #1",
			want:    http.StatusOK,
			request: "http://127.0.0.1:8080/update/",
			bodies: [][]byte{[]byte(`[{"id":"Alloc1", "type":"gauge", "value":1.0}, {"id":"Counter1", "type":"counter", "delta":1}]`),
				[]byte(`[{"id":"Alloc1", "type":"gauge", "value":2.0}, {"id":"Counter1", "type":"counter", "delta":2}]`)},
			answerGauge:   []metrics.Gauge{1.0, 2.0},
			answerCounter: []metrics.Counter{1, 3},
		},
	}
	s := storage.NewStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for idx, body := range tt.bodies {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(body))
				HandlerUpdateJSON(w, r, s, []byte{})

				require.Equal(t, http.StatusOK, w.Code)
				require.Equal(t, tt.answerGauge[idx], s.DataGauge["Alloc1"])
				require.Equal(t, tt.answerCounter[idx], s.DataCounter["Counter1"])
			}
		})
	}
}
