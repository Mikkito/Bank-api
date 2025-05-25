package handler

import (
	"bank-api/internal/middleware"
	"bank-api/internal/service"
	"bank-api/internal/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CardHandler struct {
	cardService *service.CardService
}

func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	card, err := h.cardService.CreateCard(r.Context(), userID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create card"})
		return
	}

	utils.RespondJSON(w, http.StatusCreated, card)
}

func (h *CardHandler) GetCardByID(w http.ResponseWriter, r *http.Request) {
	// userID из JWT middleware
	userIDRaw := r.Context().Value(middleware.UserIDKey)
	userID, err := strconv.ParseInt(fmt.Sprintf("%v", userIDRaw), 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// cardID из URL
	vars := mux.Vars(r)
	cardIDStr, ok := vars["id"]
	if !ok {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing card ID"})
		return
	}
	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid card ID"})
		return
	}

	cardDTO, err := h.cardService.GetDecryptedCardByID(r.Context(), userID, cardID)
	if err != nil {
		if errors.Is(err, service.ErrCardNotFound) {
			utils.RespondJSON(w, http.StatusNotFound, map[string]string{"error": "card not found"})
			return
		}
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	utils.RespondJSON(w, http.StatusOK, cardDTO)
}

func (h *CardHandler) GetAllCards(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value(middleware.UserIDKey)
	userID, err := strconv.ParseInt(fmt.Sprintf("%v", userIDRaw), 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	cards, err := h.cardService.GetCardsByUser(r.Context(), userID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch cards"})
		return
	}

	utils.RespondJSON(w, http.StatusOK, cards)
}

func (h *CardHandler) BlockCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardIDStr := vars["id"]
	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid card ID"})
		return
	}

	err = h.cardService.BlockCard(r.Context(), cardID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not block card"})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "card blocked"})
}

func (h *CardHandler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardIDStr := vars["id"]
	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid card ID"})
		return
	}

	err = h.cardService.DeleteCard(r.Context(), cardID)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not delete card"})
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "card deleted"})
}
