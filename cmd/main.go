package main

import (
	"context"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	if err := Run(ctx); err != nil {
		log.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
