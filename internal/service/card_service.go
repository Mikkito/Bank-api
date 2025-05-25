package service

import (
	"bank-api/internal/config"
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"bank-api/internal/security"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// Service for work with card
type CardService struct {
	repo      *repositories.CardRepository
	secretKey []byte
	hmacKey   []byte
}

// NewCardService created new service for cards
func NewCardService(repo *repositories.CardRepository, secretKey, hmacKey []byte) *CardService {
	return &CardService{
		repo:      repo,
		secretKey: secretKey,
		hmacKey:   hmacKey,
	}
}

// CreateCard generate new card for choosen account
func (s *CardService) CreateCard(ctx context.Context, accountID int64) (*models.Card, error) {
	cardNumber := generateCardNumber()
	cvv := generateCVV()
	expirationDate := time.Now().AddDate(3, 0, 0)

	// Формируем строку, которую будем шифровать и подписывать
	plainData := fmt.Sprintf("%s|%s", cardNumber, expirationDate.Format("01/06"))

	// Шифруем
	encryptedData, err := security.EncryptAES(plainData, s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt card data: %w", err)
	}

	// Подписываем
	signature, err := security.GenerateHMAC(plainData, s.hmacKey)
	if err != nil {
		return nil, fmt.Errorf("generate HMAC: %w", err)
	}

	// Шифруем CVV
	encryptedCVV, err := security.EncryptAES(cvv, s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt CVV: %w", err)
	}

	card := &models.Card{
		AccountID:     accountID,
		EncryptedData: encryptedData,
		HMAC:          signature,
		CVV:           encryptedCVV,
		CreatedAt:     time.Now(),
	}

	err = s.repo.CreateCard(ctx, card)
	if err != nil {
		return nil, err
	}

	return card, nil
}

// Generate card number
func generateCardNumber() string {
	return fmt.Sprintf("4%012d", randInt64(100000000000, 999999999999))
}

func generateCVV() string {
	return fmt.Sprintf("%03d", randInt64(100, 999))
}

func randInt64(min, max int64) int64 {
	return min + randInt63n(max-min+1)
}

func randInt63n(max int64) int64 {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return int64(binary.LittleEndian.Uint64(b)) % max
}

func (s *CardService) GetCardsByAccountID(ctx context.Context, accountID int64) ([]*models.Card, error) {
	cards, err := s.repo.GetCardsByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	for _, card := range cards {
		number, expiry, cvv, err := s.decryptAndVerify(card)
		if err != nil {
			// log
			continue
		}

		// Добавляем в карту, чтобы вернуть на фронт
		card.CardNumber = number
		card.ExpirationDate, _ = time.Parse("01/06", expiry)
		card.CVV = cvv
	}

	return cards, nil
}

func (s *CardService) GetCardsByUser(ctx context.Context, userID int64) ([]*models.Card, error) {
	// Получаем карты пользователя из репозитория
	cards, err := s.repo.GetCardsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cards for user %d: %w", userID, err)
	}

	// Расшифровываем данные карт и проверяем HMAC (если нужно)
	for i, card := range cards {
		number, expire, cvv, err := s.decryptAndVerify(card)
		if err != nil {
			// Можно залогировать ошибку и продолжить, если важно продолжить обработку других карт
			log.Printf("error decrypting card %d: %v", card.ID, err)
			continue
		}

		// Обновляем расшифрованные данные карты
		cards[i].CardNumber = number
		cards[i].ExpirationDate, _ = time.Parse("01/06", expire)
		cards[i].CVV = cvv
	}

	return cards, nil
}

func (s *CardService) GetCardByID(ctx context.Context, ID int64) (*models.Card, error) {
	card, err := s.repo.GetCardByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	number, expire, cvv, err := s.decryptAndVerify(card)
	if err != nil {
		// log
	}
	card.CardNumber = number
	card.ExpirationDate, _ = time.Parse("01/06", expire)
	card.CVV = cvv
	return card, nil

}

var ErrCardNotFound = errors.New("card not found")

func (s *CardService) GetDecryptedCardByID(ctx context.Context, userID, cardID int64) (*models.CardResponse, error) {
	card, err := s.repo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if card == nil || card.AccountID != userID {
		return nil, ErrCardNotFound
	}

	number, expire, _, err := s.decryptAndVerify(card)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return &models.CardResponse{
		ID:             card.ID,
		AccountID:      card.AccountID,
		CardNumber:     number,
		ExpirationDate: expire,
		CreatedAt:      card.CreatedAt,
	}, nil
}

func (s *CardService) BlockCard(ctx context.Context, cardID int64) error {
	return s.repo.SetCardStatus(ctx, cardID, "blocked") // например, если ты добавишь поле `Status` в модель
}

func (s *CardService) DeleteCard(ctx context.Context, cardID int64) error {
	return s.repo.DeleteCard(ctx, cardID)
}

func (s *CardService) decryptAndVerify(card *models.Card) (number, expire, cvv string, err error) {
	key := []byte(config.AppConfig.Encryption.Secret)

	plaintext, err := security.DecryptAES(card.EncryptedData, key)
	if err != nil {
		return "", "", "", fmt.Errorf("decrypt error: %w", err)
	}

	expectedMAC := card.HMAC
	valid, _ := security.VerifyHMAC(plaintext, expectedMAC, key)
	if !valid {
		return "", "", "", errors.New("HMAC verification failed")
	}

	// Пример: "4111111111111111|12/27|123"
	parts := strings.Split(plaintext, "|")
	if len(parts) != 3 {
		return "", "", "", errors.New("invalid decrypted format")
	}

	number = parts[0]
	expire = parts[1]
	cvv = parts[2]
	return
}
