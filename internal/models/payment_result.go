package models

type PaymentResult struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"` // "pending", "completed", "failed"
	Provider      string `json:"provider"`
	Error         string `json:"error,omitempty"` // если есть ошибка
}
