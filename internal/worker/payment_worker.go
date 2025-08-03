package worker

import (
	"context"
	"log"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/service"
	"github.com/gabreuvcr/proxy-payment/internal/util"
)

type PaymentQueueConsumer interface {
	StartQueueConsumers(n int)
}

type paymentQueueConsumer struct {
	paymentService    service.PaymentService
	defaultProcessor  service.ProcessorService
	fallbackProcessor service.ProcessorService
}

func NewPaymentQueueConsumer(
	paymentService service.PaymentService,
	defaultProcessor service.ProcessorService,
	fallbackProcessor service.ProcessorService,
) PaymentQueueConsumer {
	return &paymentQueueConsumer{
		paymentService:    paymentService,
		defaultProcessor:  defaultProcessor,
		fallbackProcessor: fallbackProcessor,
	}
}

func (w *paymentQueueConsumer) StartQueueConsumers(n int) {
	for workerID := range n {
		go w.consumeQueue(workerID)
	}
}

func (w *paymentQueueConsumer) consumeQueue(workerID int) error {
	for {
		ctx := context.Background()

		payment, err := w.paymentService.DequeuePayment()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		log.Printf("[Worker %d] Processing payment %s\n", workerID, payment.CorrelationId)

		processorUsed, err := w.processPayment(payment)
		if err != nil {
			log.Printf("[Worker %d] ERROR trying process payment: %v\n", workerID, err)
			continue
		}

		payment.ProcessedBy = processorUsed

		if err := w.paymentService.InsertPayment(ctx, &payment); err != nil {
			log.Printf("[Worker %d] ERROR trying insert into database: %v\n", workerID, err)
		} else {
			log.Printf("[Worker %d] Payment %s processed by %d\n", workerID, payment.CorrelationId, payment.ProcessedBy)
		}
	}
}

func (w *paymentQueueConsumer) processPayment(p model.Payment) (model.Processor, error) {
	return util.RandomProcessor(), nil

	// if w.defaultProcessor.IsHealthy() {
	// 	if err := w.defaultProcessor.ProcessPayment(p); err == nil {
	// 		return model.ProcessorDefault, nil
	// 	}
	// }

	// if w.fallbackProcessor.IsHealthy() {
	// 	if err := w.fallbackProcessor.ProcessPayment(p); err == nil {
	// 		return model.ProcessorFallback, nil
	// 	}
	// }

	// return model.ProcessorNone, errors.New("both processors failed")
}
