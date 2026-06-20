package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetAccrual отправляет запрос к внешнему сервису для получения информации о начислении по номеру заказа.
func (c *Client) GetAccrual(ctx context.Context, orderNumber string) (*model.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp model.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &accrualResp, nil

	case http.StatusNoContent:
		return &model.AccrualResponse{
			Order:  orderNumber,
			Status: model.AccrualStatusInvalid,
		}, nil

	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("too many requests (429)")

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
