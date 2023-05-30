package main

import (
	"github.com/luckyseadog/go-dev/pkg"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(pkg.HandlerDefault))
	mux.Handle("/update/", http.HandlerFunc(pkg.HandlerUpdate))
	server := NewServer("127.0.0.1:8080", mux)
	server.Run()
}
