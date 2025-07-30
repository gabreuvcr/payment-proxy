package main

import (
	"log"

	"github.com/gabreuvcr/proxy-payment/internal/app"
)

func main() {
	if err := app.RunServer(); err != nil {
		log.Fatal(err)
	}
}
