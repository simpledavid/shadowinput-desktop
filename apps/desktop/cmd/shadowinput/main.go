package main

import (
	"arion.local/desktop/internal/shadowinput"
	"log"
	"os"
)

func main() {
	if handled, err := shadowinput.Maintenance(os.Args[1:]); handled {
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := shadowinput.Run(); err != nil {
		log.Fatal(err)
	}
}
