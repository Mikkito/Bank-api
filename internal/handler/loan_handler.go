package handler

import (
	"bank-api/internal/middleware"
	"bank-api/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type LoanHandler struct {
	service *service.LoanService
}

type repayPartialRequest struct {
	Amount float64 `json:"amount"`
}

type takeLoanRequest struct {
	AccountID int64   `json:"account_id"`
	Amount    float64 `json:"amount"`
}

func NewLoanHandler(service *service.LoanService) *LoanHandler {
	return &LoanHandler{service: service}
}

// POST loan/take
func (h *LoanHandler) TakeLoan(w http.ResponseWriter, r *http.Request) {
	var req takeLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.TakeLoan(r.Context(), userID, req.AccountID, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET loan/
func (h *LoanHandler) GetUserLoans(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	loans, err := h.service.GetUserLoans(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loans)
}

func (h *LoanHandler) RepayLoan(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	loanID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid loan ID", http.StatusBadRequest)
		return
	}

	if err := h.service.RepayLoan(r.Context(), loanID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *LoanHandler) MarkAsRepaid(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	loanID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid loan ID", http.StatusBadRequest)
		return
	}

	if err := h.service.MarkLoanAsRepaid(r.Context(), loanID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *LoanHandler) GetOutstandingDebt(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	loanID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid loan ID", http.StatusBadRequest)
		return
	}

	loan, err := h.service.GetLoanByID(r.Context(), loanID)
	if err != nil {
		http.Error(w, "loan not found", http.StatusNotFound)
		return
	}

	debt, err := h.service.CalculateOutstandingDebt(r.Context(), loan)
	if err != nil {
		http.Error(w, "could not calculate outstanding debt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{
		"outstanding_debt": debt,
	})
}

func (h *LoanHandler) RepayPartial(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	loanID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid loan ID", http.StatusBadRequest)
		return
	}

	var req repayPartialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	err = h.service.RepayPartialLoan(r.Context(), loanID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
