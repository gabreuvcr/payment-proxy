package app

import (
	"log"
	"net/http"
	"os"

	"github.com/gabreuvcr/proxy-payment/internal/handler"
	"github.com/gabreuvcr/proxy-payment/internal/infra"
	"github.com/gabreuvcr/proxy-payment/internal/queue"
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

	paymentsQueue := queue.NewPaymentQueue(redis)
	repo := repository.NewPaymentRepository(db)

	defaultBaseUrl := os.Getenv("DEFAULT_PROCESSOR_BASE_URL")
	fallbackBaseUrl := os.Getenv("FALLBACK_PROCESSOR_BASE_URL")

	defaultProcessorService := service.NewProcessorService(defaultBaseUrl, redis, "health:default")
	fallbackProcessorService := service.NewProcessorService(fallbackBaseUrl, redis, "health:fallback")
	worker := worker.NewWorker(repo, paymentsQueue, defaultProcessorService, fallbackProcessorService)
	worker.StartWorkers(10)

	paymentService := service.NewPaymentService(repo, paymentsQueue)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	mux.HandleFunc("/payments", paymentHandler.CreatePayment)
	mux.HandleFunc("/payments-summary", paymentHandler.GetSummary)
	mux.HandleFunc("/ping", paymentHandler.Pong)

	log.Println("Server running at :9999")
	return http.ListenAndServe(":9999", mux)
}
