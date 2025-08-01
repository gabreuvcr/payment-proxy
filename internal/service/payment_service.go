package service

import (
	"context"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/queue"
	"github.com/gabreuvcr/proxy-payment/internal/repository"
)

type PaymentService interface {
	EnqueuePayment(ctx context.Context, p *model.Payment) error
	GetSummary(ctx context.Context, from *time.Time, to *time.Time) (model.PaymentSummaryResponse, error)
}

type paymentService struct {
	r repository.PaymentRepository
	q queue.Queue
}

func NewPaymentService(repo repository.PaymentRepository, queue queue.Queue) PaymentService {
	return &paymentService{
		r: repo,
		q: queue,
	}
}

func (s *paymentService) EnqueuePayment(ctx context.Context, p *model.Payment) error {
	return s.q.Enqueue(ctx, *p)
}

func (s *paymentService) GetSummary(ctx context.Context, from *time.Time, to *time.Time) (model.PaymentSummaryResponse, error) {
	return s.r.GetSummary(ctx, from, to)
}
