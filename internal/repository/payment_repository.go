package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/redis/go-redis/v9"
)

type PaymentRepository interface {
	InsertPayment(ctx context.Context, p *model.Payment) error
	GetSummary(ctx context.Context, from *time.Time, to *time.Time) (model.PaymentSummaryResponse, error)

	Enqueue(ctx context.Context, payment model.Payment) error
	Dequeue() (model.Payment, error)
	GetHealth(ctx context.Context, healthKey string) (string, error)
	SetHealthy(ctx context.Context, healthKey string)
	SetUnhealthy(ctx context.Context, healthKey string)
}

type paymentRepository struct {
	db       *sql.DB
	redis    *redis.Client
	queueKey string
}

func NewPaymentRepository(db *sql.DB, redis *redis.Client) PaymentRepository {
	return &paymentRepository{
		db:       db,
		redis:    redis,
		queueKey: "queue:payment",
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

func (q *paymentRepository) Enqueue(ctx context.Context, payment model.Payment) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	data, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	return q.redis.LPush(ctx, q.queueKey, data).Err()
}

func (q *paymentRepository) Dequeue() (model.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	data, err := q.redis.RPop(ctx, q.queueKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.Payment{}, errors.New("queue is empty")
		}
		return model.Payment{}, err
	}

	var p model.Payment
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return model.Payment{}, err
	}

	return p, nil
}

func (q *paymentRepository) GetHealth(ctx context.Context, healthKey string) (string, error) {
	return q.redis.Get(ctx, healthKey).Result()
}

func (q *paymentRepository) SetHealthy(ctx context.Context, healthKey string) {
	q.redis.Set(ctx, healthKey, "1", 5*time.Second)
}

func (q *paymentRepository) SetUnhealthy(ctx context.Context, healthKey string) {
	q.redis.Set(ctx, healthKey, "0", 5*time.Second)
}
