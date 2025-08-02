package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/redis/go-redis/v9"
)

type ProcessorService struct {
	baseURL    string
	redis      *redis.Client
	httpClient *http.Client
	healthKey  string
}

func NewProcessorService(baseURL string, redisClient *redis.Client, healthKey string) *ProcessorService {
	return &ProcessorService{
		baseURL:    baseURL,
		redis:      redisClient,
		httpClient: &http.Client{Timeout: 500 * time.Millisecond},
		healthKey:  healthKey,
	}
}

func (p *ProcessorService) IsHealthy() bool {
	ctx := context.Background()

	val, err := p.redis.Get(ctx, p.healthKey).Result()
	if err == nil {
		return val == "1"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/payments/service-health", nil)
	if err != nil {
		p.redis.Set(ctx, p.healthKey, "0", 5*time.Second)
		return false
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.redis.Set(ctx, p.healthKey, "0", 5*time.Second)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.redis.Set(ctx, p.healthKey, "0", 5*time.Second)
		return false
	}

	var serviceHealthResult struct {
		Failing         bool `json:"failing"`
		MinResponseTime int  `json:"minResponseTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&serviceHealthResult); err != nil {
		p.redis.Set(ctx, p.healthKey, "0", 5*time.Second)
		return false
	}

	if serviceHealthResult.Failing {
		p.redis.Set(ctx, p.healthKey, "0", 5*time.Second)
		return false
	}

	p.redis.Set(ctx, p.healthKey, "1", 5*time.Second)
	return true
}

func (p *ProcessorService) ProcessPayment(payment model.Payment) error {
	ctx := context.Background()

	data, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/payments", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("processor at %s returned status %d", p.baseURL, resp.StatusCode)
	}

	return nil
}
