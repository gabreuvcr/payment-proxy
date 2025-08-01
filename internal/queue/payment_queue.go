package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/redis/go-redis/v9"
)

type Queue interface {
	Enqueue(ctx context.Context, payment model.Payment) error
	Dequeue() (model.Payment, error)
}

type PaymentQueue struct {
	client *redis.Client
	key    string
}

func NewPaymentQueue(client *redis.Client) *PaymentQueue {
	return &PaymentQueue{
		client: client,
		key:    "payment-queue",
	}
}

func (q *PaymentQueue) Enqueue(ctx context.Context, payment model.Payment) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	data, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	return q.client.LPush(ctx, q.key, data).Err()
}

func (q *PaymentQueue) Dequeue() (model.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	data, err := q.client.RPop(ctx, q.key).Result()
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
