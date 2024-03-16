package main

import (
	"fmt"
	"os"

	"github.com/marcozac/go-aliaser/cmd/aliaser/internal"
)

func main() {
	if err := internal.NewRoot().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
