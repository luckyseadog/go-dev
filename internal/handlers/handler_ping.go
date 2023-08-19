package handlers

import (
	"net/http"

	"github.com/luckyseadog/go-dev/internal/storage"
)

// HandlerPing is an HTTP handler that responds to GET requests with a simple "COOL!" message if the SQL storage's
// database is reachable. The function checks for the availability of the database by sending a ping request to it.
// If the database is not available or if the provided storage is not of type *storage.SQLStorage, the function
// responds with an appropriate HTTP error message.
//
// Parameters:
//   - w: The http.ResponseWriter to write the HTTP response.
//   - r: The http.Request received from the client.
//   - s: An instance of storage.Storage used to check the database availability.
//
// Notes:
//   - The function requires the provided storage to be of type *storage.SQLStorage for proper execution.
func HandlerPing(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	if r.Method != http.MethodGet {
		http.Error(w, "HandlerPing: Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	ss, ok := s.(*storage.SQLStorage)
	if !ok {
		http.Error(w, "HandlerPing: this method allowed only with SQLStorage", http.StatusMethodNotAllowed)
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
