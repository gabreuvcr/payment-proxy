package model

import (
	"time"
)

type Processor int16

const (
	ProcessorDefault  Processor = 0
	ProcessorFallback Processor = 1
)

type Payment struct {
	CorrelationId string
	Amount        float64
	ProcessedBy   Processor
	RequestedAt   time.Time
}

type PaymentSummaryDetails struct {
	ProcessedBy   Processor
	TotalRequests int64
	TotalAmount   float64
}

type CreatePaymentRequest struct {
	CorrelationId string  `json:"correlationId" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
}

type PaymentSummaryDetailsResponse struct {
	TotalRequests int64   `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentSummaryResponse struct {
	Default  PaymentSummaryDetailsResponse `json:"default"`
	Fallback PaymentSummaryDetailsResponse `json:"fallback"`
}
