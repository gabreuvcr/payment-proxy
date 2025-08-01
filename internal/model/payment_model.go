package model

import (
	"time"
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
