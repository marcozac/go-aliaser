package main

import (
	"log"

	"github.com/marcozac/go-aliaser"
)

// main generates the aliases for the gin package and writes them to gin/alias.go.
func main() {
	a, err := aliaser.New("gin", "github.com/gin-gonic/gin")
	if err != nil {
		log.Fatal(err)
	}
	if err := a.GenerateFile("gin/alias.go"); err != nil {
		log.Fatal(err)
	}
}
