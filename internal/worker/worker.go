package worker

import (
	"context"
	"log"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/service"
	"github.com/gabreuvcr/proxy-payment/internal/util"
)

type Worker struct {
	paymentService    service.PaymentService
	defaultProcessor  service.ProcessorService
	fallbackProcessor service.ProcessorService
}

func NewWorker(
	paymentService service.PaymentService,
	defaultProc service.ProcessorService,
	fallbackProc service.ProcessorService,
) *Worker {
	return &Worker{
		paymentService:    paymentService,
		defaultProcessor:  defaultProc,
		fallbackProcessor: fallbackProc,
	}
}

func (w *Worker) StartWorkers(n int) {
	for i := range n {
		go w.workerLoop(i)
	}
}

func (w *Worker) workerLoop(id int) error {
	for {
		ctx := context.Background()

		payment, err := w.paymentService.DequeuePayment()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		log.Printf("[Worker %d] Processing payment %s\n", id, payment.CorrelationId)

		processorUsed, err := w.processPayment(payment)
		if err != nil {
			log.Printf("[Worker %d] ERROR trying process payment: %v\n", id, err)
			continue
		}

		payment.ProcessedBy = processorUsed

		if err := w.paymentService.InsertPayment(ctx, &payment); err != nil {
			log.Printf("[Worker %d] ERROR trying insert into database: %v\n", id, err)
		} else {
			log.Printf("[Worker %d] Payment %s processed by %d\n", id, payment.CorrelationId, payment.ProcessedBy)
		}
	}
}

func (w *Worker) processPayment(p model.Payment) (model.Processor, error) {
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
