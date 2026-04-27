package main

import (
	"log"
	"os"

	"github.com/kainnsoft/duoscans/internal/builder"
)

func main() {
	if err := builder.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
