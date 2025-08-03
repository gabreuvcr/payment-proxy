package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/repository"
)

type ProcessorService interface {
	IsHealthy() bool
	ProcessPayment(payment model.Payment) error
}

type processorService struct {
	baseURL    string
	repo       repository.PaymentRepository
	httpClient *http.Client
	healthKey  string
}

func NewProcessorService(baseURL string, repo repository.PaymentRepository, healthKey string) ProcessorService {
	return &processorService{
		baseURL:    baseURL,
		repo:       repo,
		httpClient: &http.Client{Timeout: 500 * time.Millisecond},
		healthKey:  healthKey,
	}
}

func (p *processorService) IsHealthy() bool {
	ctx := context.Background()

	val, err := p.repo.GetHealth(ctx, p.healthKey)
	if err == nil {
		return val == "1"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/payments/service-health", nil)
	if err != nil {
		p.repo.SetUnhealthy(ctx, p.healthKey)
		return false
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.repo.SetUnhealthy(ctx, p.healthKey)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.repo.SetUnhealthy(ctx, p.healthKey)
		return false
	}

	var serviceHealthResult struct {
		Failing         bool `json:"failing"`
		MinResponseTime int  `json:"minResponseTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&serviceHealthResult); err != nil {
		p.repo.SetUnhealthy(ctx, p.healthKey)
		return false
	}

	if serviceHealthResult.Failing {
		p.repo.SetUnhealthy(ctx, p.healthKey)
		return false
	}

	p.repo.SetHealthy(ctx, p.healthKey)
	return true
}

func (p *processorService) ProcessPayment(payment model.Payment) error {
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("processor at %s returned status %d", p.baseURL, resp.StatusCode)
	}

	return nil
}
