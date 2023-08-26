package notmain

import (
	"log"
	"os"
)

func main() {
	if true {
		os.Exit(1)
	} else {
		log.Fatal("error")
	}
}
