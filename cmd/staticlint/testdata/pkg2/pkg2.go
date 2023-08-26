package main

import (
	"log"
	"os"
)

func someFunc() {
	if true {
		os.Exit(1)
	} else {
		log.Fatal("error")
	}
}
