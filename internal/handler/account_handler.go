package handler

import (
	"bank-api/internal/service"
	"bank-api/internal/utils"
	"context"
	"errors"
	"net/http"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	// Получение userID из контекста (установлено в middleware)
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	account, err := h.accountService.CreateAccount(context.Background(), userID)
	if err != nil {
		if errors.Is(err, service.ErrAccountAlreadyExists) {
			utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "account already exists"})
			return
		}
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	utils.RespondJSON(w, http.StatusCreated, account)
}
