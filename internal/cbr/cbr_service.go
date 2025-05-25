package cbr

import (
	"context"
	"fmt"
	"net/http"

	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type CBRService struct {
	client *http.Client
}

func NewCBRService(client *http.Client) *CBRService {
	return &CBRService{client: client}
}

func (s *CBRService) GetKeyRate(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://cbr.ru/", nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}

	var rate float64
	doc.Find(".key-indicator__value").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		text = strings.Replace(text, ",", ".", 1)
		rate, err = strconv.ParseFloat(text, 64)
		return err != nil // остановить перебор при успешном разборе
	})

	if err != nil {
		return 0, err
	}

	return rate, nil
}
