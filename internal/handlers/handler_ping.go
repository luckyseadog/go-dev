package handlers

import (
	"github.com/luckyseadog/go-dev/internal/storage"
	"net/http"
)

func HandlerPing(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	if r.Method != http.MethodGet {
		http.Error(w, "HandlerDefault: Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	ss, ok := s.(*storage.SqlStorage)
	if !ok {
		http.Error(w, "Configuration Error: this method allowed only with SqlStorage", http.StatusMethodNotAllowed)
		return
	}

	err := ss.DB.Ping()
	if err != nil {
		http.Error(w, "HandlerPing: DataBase is not available", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("COOL!"))
}
