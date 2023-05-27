package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	mainHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, req.URL.String())
		fmt.Println(req.URL.String())
	}

	http.HandleFunc("/", mainHandler)
	log.Fatal(http.ListenAndServe(":8083", nil))
}
