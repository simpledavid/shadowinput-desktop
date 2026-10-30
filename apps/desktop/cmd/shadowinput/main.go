package main

import (
	"arion.local/desktop/internal/shadowinput"
	"log"
)

func main() {
	if err := shadowinput.Run(); err != nil {
		log.Fatal(err)
	}
}
