package handler

import (
	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"
	"bank-api/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// POST /payments
func (h *PaymentHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	var method models.PaymentMethod
	if err := json.NewDecoder(r.Body).Decode(&method); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	method.UserID = userID

	if err := h.paymentService.AddPaymentMethod(ctx, &method); err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusCreated, method)
}

// GET /payments
func (h *PaymentHandler) GetUserMethods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	methods, err := h.paymentService.GetActiveMethodsByUser(ctx, userID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, methods)
}

// DELETE /payments/{id}
func (h *PaymentHandler) DeactivateMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid method ID"})
		return
	}

	if err := h.paymentService.DeactivateMethod(ctx, id); err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}
