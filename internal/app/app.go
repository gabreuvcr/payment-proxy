package app

import (
	"log"
	"net/http"
	"os"

	"github.com/gabreuvcr/proxy-payment/internal/handler"
	"github.com/gabreuvcr/proxy-payment/internal/infra"
	"github.com/gabreuvcr/proxy-payment/internal/repository"
	"github.com/gabreuvcr/proxy-payment/internal/service"
	"github.com/gabreuvcr/proxy-payment/internal/worker"
)

func Run() error {
	mux := http.NewServeMux()

	db, err := infra.NewDb()
	if err != nil {
		return err
	}
	redis := infra.NewRedis()

	repo := repository.NewPaymentRepository(db, redis)

	defaultBaseUrl := os.Getenv("DEFAULT_PROCESSOR_BASE_URL")
	fallbackBaseUrl := os.Getenv("FALLBACK_PROCESSOR_BASE_URL")

	defaultProcessorService := service.NewProcessorService(defaultBaseUrl, repo, "health:default")
	fallbackProcessorService := service.NewProcessorService(fallbackBaseUrl, repo, "health:fallback")
	paymentService := service.NewPaymentService(repo)

	paymentHandler := handler.NewPaymentHandler(paymentService)

	worker := worker.NewPaymentQueueConsumer(paymentService, defaultProcessorService, fallbackProcessorService)
	worker.StartQueueConsumers(10)

	mux.HandleFunc("/payments", paymentHandler.CreatePayment)
	mux.HandleFunc("/payments-summary", paymentHandler.GetSummary)
	mux.HandleFunc("/ping", paymentHandler.Pong)

	log.Println("Server running at :9999")
	return http.ListenAndServe(":9999", mux)
}
