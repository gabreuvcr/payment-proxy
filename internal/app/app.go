package app

import (
	"log"
	"net/http"

	"github.com/gabreuvcr/proxy-payment/internal/config"
	"github.com/gabreuvcr/proxy-payment/internal/handler"
	"github.com/gabreuvcr/proxy-payment/internal/repository"
	"github.com/gabreuvcr/proxy-payment/internal/service"
)

func RunServer() error {
	mux := http.NewServeMux()

	db, err := config.NewDb()
	if err != nil {
		return err
	}

	repo := repository.NewPaymentRepository(db)
	service := service.NewPaymentService(repo)
	handler := handler.NewPaymentHandler(service)

	mux.HandleFunc("/payments", handler.CreatePayment) // corrigir path aqui: sem verbo HTTP no HandleFunc
	mux.HandleFunc("/payments-summary", handler.GetSummary)

	log.Println("Server running at :9999")
	return http.ListenAndServe(":9999", mux)
}
