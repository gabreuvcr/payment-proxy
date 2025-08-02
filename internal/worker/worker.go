package worker

import (
	"context"
	"log"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/queue"
	"github.com/gabreuvcr/proxy-payment/internal/repository"
	"github.com/gabreuvcr/proxy-payment/internal/service"
	"github.com/gabreuvcr/proxy-payment/internal/util"
)

type Worker struct {
	repo              repository.PaymentRepository
	queue             queue.Queue
	defaultProcessor  *service.ProcessorService
	fallbackProcessor *service.ProcessorService
}

func NewWorker(
	repo repository.PaymentRepository,
	queue queue.Queue,
	defaultProc *service.ProcessorService,
	fallbackProc *service.ProcessorService,
) *Worker {
	return &Worker{
		repo:              repo,
		queue:             queue,
		defaultProcessor:  defaultProc,
		fallbackProcessor: fallbackProc,
	}
}

func (s *Worker) StartWorkers(n int) {
	for i := range n {
		go s.workerLoop(i)
	}
}

func (s *Worker) workerLoop(id int) error {
	for {
		ctx := context.Background()

		payment, err := s.queue.Dequeue()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		log.Printf("[Worker %d] Processing payment %s\n", id, payment.CorrelationId)

		processorUsed, err := s.processPayment(payment)
		if err != nil {
			log.Printf("[Worker %d] ERROR trying process payment: %v\n", id, err)
			continue
		}

		payment.ProcessedBy = processorUsed

		if err := s.repo.InsertPayment(ctx, &payment); err != nil {
			log.Printf("[Worker %d] ERROR trying insert into database: %v\n", id, err)
		} else {
			log.Printf("[Worker %d] Payment %s processed by %d\n", id, payment.CorrelationId, payment.ProcessedBy)
		}
	}
}

func (s *Worker) processPayment(p model.Payment) (model.Processor, error) {
	return util.RandomProcessor(), nil

	// if s.defaultProcessor.IsHealthy() {
	// 	if err := s.defaultProcessor.ProcessPayment(p); err == nil {
	// 		return model.ProcessorDefault, nil
	// 	}
	// }

	// if s.fallbackProcessor.IsHealthy() {
	// 	if err := s.fallbackProcessor.ProcessPayment(p); err == nil {
	// 		return model.ProcessorFallback, nil
	// 	}
	// }

	// return model.ProcessorNone, errors.New("both processors failed")
}
