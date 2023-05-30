package pkg

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			want:    http.StatusNotFound,
			request: "http://127.0.0.1:8080/update/hh/Alloc/1.0",
		},
		{
			name:    "simple test #3",
			want:    http.StatusInternalServerError,
			request: "http://127.0.0.1:8080/update/gauge/Alloc/hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(HandlerUpdate)
			h(w, request)
			result := w.Result()
			assert.Equal(t, tt.want, result.StatusCode)
			defer result.Body.Close()
			_, err := io.Copy(io.Discard, result.Body)
			assert.NoError(t, err)
		})
	}
}
