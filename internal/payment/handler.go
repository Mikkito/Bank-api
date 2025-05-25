package payment

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/utils"

	"github.com/gorilla/mux"
)

type PaymentHandler struct {
	paymentService *PaymentService
}

func NewPaymentHandler(paymentService *PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// POST /payment-methods
func (h *PaymentHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	var input models.PaymentMethod
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input"})
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	input.UserID = userID

	if err := h.paymentService.AddPaymentMethod(r.Context(), &input); err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusCreated, input)
}

// GET /payment-methods
func (h *PaymentHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	methods, err := h.paymentService.GetPaymentMethods(r.Context(), userID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, methods)
}

// DELETE /payment-methods/{id}
func (h *PaymentHandler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payment method ID"})
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.paymentService.DeletePaymentMethod(r.Context(), userID, id); err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var payment models.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input"})
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	payment.UserID = userID

	result, err := h.paymentService.ProcessPayment(r.Context(), payment)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, result)
}

// POST /payments/{id}/refund
func (h *PaymentHandler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	paymentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payment ID"})
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	payment, err := h.paymentService.GetPaymentByID(r.Context(), paymentID)
	if err != nil {
		utils.RespondJSON(w, http.StatusNotFound, map[string]string{"error": "payment not found"})
		return
	}

	if payment.UserID != userID {
		utils.RespondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	if err := h.paymentService.RefundPayment(r.Context(), paymentID); err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "refunded"})
}
