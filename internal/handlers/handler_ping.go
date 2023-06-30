package handlers

import (
	"database/sql"
	"net/http"
)

func HandlerPing(w http.ResponseWriter, r *http.Request, storage *sql.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "HandlerDefault: Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	err := storage.Ping()
	if err != nil {
		http.Error(w, "HandlerPing: DataBase is not available", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("COOL!"))
}
