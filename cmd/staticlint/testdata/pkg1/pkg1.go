package main

import (
	"log"
	"os"
)

func main() {
	if true {
		os.Exit(1) // want "os.Exit is prohibited in main"
	} else {
		log.Fatal("error")
	}
}
