package main

import (
	"log"

	"ecommerce_go/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
