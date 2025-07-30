package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/service"
	"github.com/gabreuvcr/proxy-payment/internal/util"
)

type PaymentHandler struct {
	s service.PaymentService
}

func NewPaymentHandler(s service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		s: s,
	}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var paymentDto model.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentDto); err != nil {
		log.Println("Error [json.NewDecoder.Decode]: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payment = model.Payment{
		CorrelationId: paymentDto.CorrelationId,
		Amount:        paymentDto.Amount,
		ProcessedBy:   util.RandomProcessor(),
		RequestedAt:   time.Now().UTC(),
	}

	if err := h.s.CreatePayment(r.Context(), &payment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PaymentHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	queryFrom := r.URL.Query().Get("from")
	queryTo := r.URL.Query().Get("to")

	var from, to *time.Time

	if queryFrom != "" {
		t, err := time.Parse(time.RFC3339, queryFrom)
		if err != nil {
			http.Error(w, "Invalid 'from' date format. Use RFC3339.", http.StatusBadRequest)
			return
		}
		from = &t
	}

	if queryTo != "" {
		t, err := time.Parse(time.RFC3339, queryTo)
		if err != nil {
			http.Error(w, "Invalid 'to' date format. Use RFC3339.", http.StatusBadRequest)
			return
		}
		to = &t
	}

	response, err := h.s.GetSummary(r.Context(), from, to)
	if err != nil {
		http.Error(w, "Failed to get summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
