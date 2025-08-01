package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
)

type PaymentRepository interface {
	InsertPayment(ctx context.Context, p *model.Payment) error
	GetSummary(ctx context.Context, from *time.Time, to *time.Time) (model.PaymentSummaryResponse, error)
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{
		db: db,
	}
}

func (r *paymentRepository) InsertPayment(ctx context.Context, p *model.Payment) error {
	query := `
		INSERT INTO payments (correlation_id, amount, processed_by, requested_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.CorrelationId,
		p.Amount,
		p.ProcessedBy,
		p.RequestedAt,
	)
	if err != nil {
		log.Println("Error [r.db.ExecContext]: ", err)
	}
	return err
}

func (r *paymentRepository) GetSummary(ctx context.Context, from *time.Time, to *time.Time) (model.PaymentSummaryResponse, error) {
	query := `
		SELECT
			processed_by AS ProcessedBy,
			COUNT(*) AS TotalRequests,
			SUM(amount) AS TotalAmount
		FROM payments
		WHERE
		($1::timestamp IS NULL OR requested_at >= $1)
		AND ($2::timestamp IS NULL OR requested_at <= $2)
		GROUP BY processed_by;
	`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		log.Println("Error [r.db.QueryContext]: ", err)
		return model.PaymentSummaryResponse{}, err
	}

	summaries := []model.PaymentSummaryDetails{}

	for rows.Next() {
		summary := model.PaymentSummaryDetails{}
		rows.Scan(&summary.ProcessedBy, &summary.TotalRequests, &summary.TotalAmount)
		summaries = append(summaries, summary)
	}

	summaryResponse := model.PaymentSummaryResponse{
		Default: model.PaymentSummaryDetailsResponse{
			TotalRequests: 0,
			TotalAmount:   0,
		},
		Fallback: model.PaymentSummaryDetailsResponse{
			TotalRequests: 0,
			TotalAmount:   0,
		},
	}

	for _, summary := range summaries {
		detail := model.PaymentSummaryDetailsResponse{
			TotalRequests: summary.TotalRequests,
			TotalAmount:   summary.TotalAmount,
		}
		switch summary.ProcessedBy {
		case model.ProcessorDefault:
			summaryResponse.Default = detail
		case model.ProcessorFallback:
			summaryResponse.Fallback = detail
		}
	}
	return summaryResponse, nil
}
