package main

import (
	"log"

	"github.com/andreygolubkow/we-know/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("❌ error: %v", err)
	}
}
