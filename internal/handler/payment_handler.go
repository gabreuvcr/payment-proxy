package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gabreuvcr/proxy-payment/internal/model"
	"github.com/gabreuvcr/proxy-payment/internal/service"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var paymentDto model.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentDto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payment = model.Payment{
		CorrelationId: paymentDto.CorrelationId,
		Amount:        paymentDto.Amount,
		RequestedAt:   time.Now().UTC(),
	}

	if err := h.s.EnqueuePayment(r.Context(), &payment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PaymentHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	queryFrom := r.URL.Query().Get("from")
	queryTo := r.URL.Query().Get("to")

	var from, to *time.Time

	if queryFrom != "" {
		if t, err := time.Parse(time.RFC3339, queryFrom); err == nil {
			from = &t
		} else {
			http.Error(w, "Invalid 'from' date format. Use RFC3339.", http.StatusBadRequest)
			return
		}
	}

	if queryTo != "" {
		if t, err := time.Parse(time.RFC3339, queryTo); err == nil {
			to = &t
		} else {
			http.Error(w, "Invalid 'to' date format. Use RFC3339.", http.StatusBadRequest)
			return
		}
	}

	response, err := h.s.GetSummary(r.Context(), from, to)
	if err != nil {
		http.Error(w, "Failed to get summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PaymentHandler) Pong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}
