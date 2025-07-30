package main

import (
	"log"
	"net/http"

	"github.com/gabreuvcr/proxy-payment/internal/config"
	"github.com/gabreuvcr/proxy-payment/internal/handler"
	"github.com/gabreuvcr/proxy-payment/internal/repository"
	"github.com/gabreuvcr/proxy-payment/internal/service"
)

func main() {
	mux := http.NewServeMux()

	db, err := config.NewDb()
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPaymentRepository(db)
	service := service.NewPaymentService(repo)
	handler := handler.NewPaymentHandler(service)

	mux.HandleFunc("POST /payments", handler.CreatePayment)
	mux.HandleFunc("GET /payments-summary", handler.GetSummary)

	log.Println("Server running at :9999")
	log.Fatal(http.ListenAndServe(":9999", mux))
}
