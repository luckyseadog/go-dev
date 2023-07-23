package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"
)

func setupRoutes(s storage.Storage) *chi.Mux {
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
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			HandlerUpdatesJSON(w, r, s, []byte{})
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			HandlerUpdatesJSON(w, r, s, []byte{})
		})
	})

	return r
}

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
	s := storage.NewStorage(nil, time.Second)
	r := setupRoutes(s)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := http.Post(ts.URL+tt.request, "text/plain", nil)
			require.NoError(t, err)
			defer response.Body.Close()

			require.Equal(t, http.StatusOK, response.StatusCode)
			_, err = io.Copy(io.Discard, response.Body)
			require.NoError(t, err)

			response, err = http.Get(ts.URL)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(body))
		})
	}
}

func TestHandlerGet(t *testing.T) {
	tests := []struct {
		name    string
		request string
		want    int
	}{
		{
			name:    "TestHandlerGet #1",
			request: "/value/gauge/Malloc",
			want:    http.StatusOK,
		},
		{
			name:    "TestHandlerGet #2",
			request: "/value/counter/Counter",
			want:    http.StatusOK,
		},
		{
			name:    "TestHandlerGet #3",
			request: "/value/counter/unknown",
			want:    http.StatusNotFound,
		},
	}

	s := storage.NewStorage(nil, time.Second)
	s.DataGauge = map[metrics.Metric]metrics.Gauge{
		"Malloc": 10.0,
	}
	s.DataCounter = map[metrics.Metric]metrics.Counter{
		"Counter": 10,
	}
	r := setupRoutes(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			require.Equal(t, tt.want, result.StatusCode)
			defer result.Body.Close()

			_, err := io.Copy(io.Discard, result.Body)
			require.NoError(t, err)
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
	s := storage.NewStorage(nil, time.Second)
	r := setupRoutes(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			require.Equal(t, tt.want, result.StatusCode)
			defer result.Body.Close()

			_, err := io.Copy(io.Discard, result.Body)
			require.NoError(t, err)
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
			bodies: [][]byte{[]byte(`{"id":"Alloc1", "type":"gauge", "value":1.0}`),
				[]byte(`{"id":"Counter1", "type":"counter", "delta":2}`),
				[]byte(`{"id":"Counter1", "type":"counter", "delta":2}`)},
			answerGauge:   []metrics.Gauge{1.0, 1.0, 1.0},
			answerCounter: []metrics.Counter{0, 2, 4},
		},
	}
	s := storage.NewStorage(nil, time.Second)
	r := setupRoutes(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for idx, body := range tt.bodies {
				w := httptest.NewRecorder()
				request := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(body))

				r.ServeHTTP(w, request)
				result := w.Result()
				require.Equal(t, tt.want, result.StatusCode)
				defer result.Body.Close()

				_, err := io.Copy(io.Discard, result.Body)
				require.NoError(t, err)
				require.Equal(t, tt.answerGauge[idx], s.DataGauge["Alloc1"])
				require.Equal(t, tt.answerCounter[idx], s.DataCounter["Counter1"])
			}
		})
	}
}

func TestHandlerUpdatesJSON(t *testing.T) {
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
			request: "http://127.0.0.1:8080/updates/",
			bodies: [][]byte{[]byte(`[{"id":"Alloc1", "type":"gauge", "value":1.0}, {"id":"Counter1", "type":"counter", "delta":1}]`),
				[]byte(`[{"id":"Alloc1", "type":"gauge", "value":2.0}, {"id":"Counter1", "type":"counter", "delta":2}]`)},
			answerGauge:   []metrics.Gauge{1.0, 2.0},
			answerCounter: []metrics.Counter{1, 3},
		},
	}
	s := storage.NewStorage(nil, time.Second)
	r := setupRoutes(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for idx, body := range tt.bodies {
				w := httptest.NewRecorder()
				request := httptest.NewRequest("POST", "/updates/", bytes.NewBuffer(body))

				r.ServeHTTP(w, request)
				result := w.Result()
				require.Equal(t, tt.want, result.StatusCode)
				defer result.Body.Close()

				_, err := io.Copy(io.Discard, result.Body)
				require.NoError(t, err)
				require.Equal(t, tt.answerGauge[idx], s.DataGauge["Alloc1"])
				require.Equal(t, tt.answerCounter[idx], s.DataCounter["Counter1"])
			}
		})
	}
}

func TestHandlerValueJSON(t *testing.T) {
	tests := []struct {
		name    string
		request string
		want    int
		body    []metrics.Metrics
	}{
		{
			name:    "TestHandlerValueJSON #1",
			request: "/value",
			want:    http.StatusOK,
			body: []metrics.Metrics{
				{
					ID:    "Malloc",
					MType: "gauge",
				},
			},
		},
		{
			name:    "TestHandlerValueJSON #2",
			request: "/value",
			want:    http.StatusOK,
			body: []metrics.Metrics{
				{
					ID:    "Malloc",
					MType: "gauge",
				},
				{
					ID:    "Counter",
					MType: "counter",
				},
			},
		},
		{
			name:    "TestHandlerValueJSON #3",
			request: "/value",
			want:    http.StatusNotFound,
			body: []metrics.Metrics{
				{
					ID:    "Malloc",
					MType: "gauge",
				},
				{
					ID:    "Counter",
					MType: "counter",
				},
				{
					ID:    "unknown",
					MType: "counter",
				},
			},
		},
	}
	s := storage.NewStorage(nil, time.Second)
	s.DataGauge = map[metrics.Metric]metrics.Gauge{
		"Malloc": 10.0,
	}
	s.DataCounter = map[metrics.Metric]metrics.Counter{
		"Counter": 10,
	}
	r := setupRoutes(s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, tt.request, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			require.Equal(t, tt.want, result.StatusCode)
			defer result.Body.Close()

			_, err = io.Copy(io.Discard, result.Body)
			require.NoError(t, err)
		})
	}

}
