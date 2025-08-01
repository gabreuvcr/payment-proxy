package main

import (
	"log"

	"github.com/gabreuvcr/proxy-payment/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
