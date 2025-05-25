package handler

import (
	"bank-api/internal/middleware"
	"bank-api/internal/repositories"
	"bank-api/internal/service"
	"bank-api/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
	transactionRepo    *repositories.TransactionRepository
	accountRepo        *repositories.AccountRepository
}

func NewTransactionHandler(ts *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: ts}
}

type TransactionRequest struct {
	FromAccountID int64   `json:"from_account,omitempty"`
	ToAccountID   int64   `json:"to_account,omitempty"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description,omitempty"`
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.transactionService.Transfer(r.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type ReversalRequest struct {
	TransactionID int64  `json:"transaction_id"`
	Description   string `json:"description"`
}

func (h *TransactionHandler) ReverseTransaction(w http.ResponseWriter, r *http.Request) {
	var req ReversalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.transactionService.ReverseTransaction(r.Context(), req.TransactionID, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TransactionHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.transactionService.Deposit(r.Context(), req.ToAccountID, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TransactionHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.transactionService.Withdraw(r.Context(), req.FromAccountID, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TransactionHandler) CreditPayment(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.transactionService.CreditPayment(r.Context(), req.FromAccountID, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TransactionHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Получаем userID из контекста
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// 2. Получаем accountID из query параметра
	accountIDStr := r.URL.Query().Get("account_id")
	if accountIDStr == "" {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "account_id is required"})
		return
	}
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid account_id"})
		return
	}

	// 3. Проверка, что аккаунт принадлежит пользователю
	ok, err := h.accountRepo.IsAccountOwnedByUser(ctx, accountID, userID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to verify account ownership"})
		return
	}
	if !ok {
		utils.RespondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied to this account"})
		return
	}

	// 4. Получаем историю транзакций
	transactions, err := h.transactionRepo.GetTransactionsByAccount(ctx, accountID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch transactions"})
		return
	}

	// 5. Отдаем результат
	utils.RespondJSON(w, http.StatusOK, transactions)
}
