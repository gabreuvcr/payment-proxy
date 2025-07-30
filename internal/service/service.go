package service

import (
	"context"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/repository"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, p *model.Payment) error
	GetSummary(ctx context.Context, from *time.Time, to *time.Time) (model.PaymentSummaryResponse, error)
}

type paymentService struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) PaymentService {
	return &paymentService{
		repo: repo,
	}
}

func (s *paymentService) CreatePayment(ctx context.Context, p *model.Payment) error {
	return s.repo.InsertPayment(ctx, p)
}

func (s *paymentService) GetSummary(ctx context.Context, from *time.Time, to *time.Time) (model.PaymentSummaryResponse, error) {
	return s.repo.GetSummary(ctx, from, to)
}
