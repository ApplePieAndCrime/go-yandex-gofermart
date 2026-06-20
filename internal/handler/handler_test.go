package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/auth"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/middleware"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/service"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func initTestJWT(t *testing.T) {
	t.Helper()
	if err := os.Setenv("JWT_SECRET", "test-secret"); err != nil {
		t.Fatal(err)
	}
	if err := auth.InitJWT(); err != nil {
		t.Fatal(err)
	}
}

func ptr(f float64) *float64 {
	return &f
}

type mockRepo struct {
	createUserFunc          func(ctx context.Context, login, hashed string) (int, error)
	getUserByLoginFunc      func(ctx context.Context, login string) (*model.User, error)
	getUserByIDFunc         func(ctx context.Context, userID int) (*model.User, error)
	insertOrderFunc         func(ctx context.Context, number string, userID int) (bool, error)
	getOrderByNumberFunc    func(ctx context.Context, number string) (*model.Order, error)
	updateOrderStatusFunc   func(ctx context.Context, number string, status string, accrual *float64) error
	getUserOrdersFunc       func(ctx context.Context, userID int) ([]model.Order, error)
	getUserBalanceFunc      func(ctx context.Context, userID int) (float64, float64, error)
	withdrawFunc            func(ctx context.Context, userID int, orderNumber string, sum float64) error
	getUserWithdrawalsFunc  func(ctx context.Context, userID int) ([]model.Withdrawal, error)
	getPendingOrdersFunc    func(ctx context.Context, limit int) ([]model.Order, error)
	processOrderAccrualFunc func(ctx context.Context, orderNumber string, userID int, status model.OrderStatus, accrual *float64) error
}

func (m *mockRepo) CreateUser(ctx context.Context, login, hashed string) (int, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, login, hashed)
	}
	return 0, nil
}
func (m *mockRepo) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	if m.getUserByLoginFunc != nil {
		return m.getUserByLoginFunc(ctx, login)
	}
	return nil, nil
}
func (m *mockRepo) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockRepo) InsertOrder(ctx context.Context, number string, userID int) (bool, error) {
	if m.insertOrderFunc != nil {
		return m.insertOrderFunc(ctx, number, userID)
	}
	return true, nil
}
func (m *mockRepo) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	if m.getOrderByNumberFunc != nil {
		return m.getOrderByNumberFunc(ctx, number)
	}
	return nil, nil
}
func (m *mockRepo) UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error {
	if m.updateOrderStatusFunc != nil {
		return m.updateOrderStatusFunc(ctx, number, status, accrual)
	}
	return nil
}
func (m *mockRepo) GetUserOrders(ctx context.Context, userID int) ([]model.Order, error) {
	if m.getUserOrdersFunc != nil {
		return m.getUserOrdersFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockRepo) GetUserBalance(ctx context.Context, userID int) (float64, float64, error) {
	if m.getUserBalanceFunc != nil {
		return m.getUserBalanceFunc(ctx, userID)
	}
	return 0, 0, nil
}
func (m *mockRepo) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	if m.withdrawFunc != nil {
		return m.withdrawFunc(ctx, userID, orderNumber, sum)
	}
	return nil
}
func (m *mockRepo) GetUserWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	if m.getUserWithdrawalsFunc != nil {
		return m.getUserWithdrawalsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockRepo) GetPendingOrders(ctx context.Context, limit int) ([]model.Order, error) {
	if m.getPendingOrdersFunc != nil {
		return m.getPendingOrdersFunc(ctx, limit)
	}
	return nil, nil
}
func (m *mockRepo) ProcessOrderAccrual(ctx context.Context, orderNumber string, userID int, status model.OrderStatus, accrual *float64) error {
	if m.processOrderAccrualFunc != nil {
		return m.processOrderAccrualFunc(ctx, orderNumber, userID, status, accrual)
	}
	return nil
}

func TestHandler_Register(t *testing.T) {
	initTestJWT(t)
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name       string
		body       string
		mockCreate func(ctx context.Context, login, hashed string) (int, error)
		wantCode   int
		wantAuth   bool
	}{
		{
			name: "success",
			body: `{"login":"newuser","password":"pass"}`,
			mockCreate: func(ctx context.Context, login, hashed string) (int, error) {
				return 1, nil
			},
			wantCode: http.StatusOK,
			wantAuth: true,
		},
		{
			name: "conflict",
			body: `{"login":"existing","password":"pass"}`,
			mockCreate: func(ctx context.Context, login, hashed string) (int, error) {
				return 0, repository.ErrLoginAlreadyExists
			},
			wantCode: http.StatusConflict,
			wantAuth: false,
		},
		{
			name:       "invalid json",
			body:       `{"login":"test"`,
			mockCreate: nil,
			wantCode:   http.StatusBadRequest,
			wantAuth:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{createUserFunc: tt.mockCreate}
			svc := service.NewService(repo, "", logger)
			h := NewHandler(svc, logger)

			router := chi.NewRouter()
			router.Post("/api/user/register", h.Register)

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
			if tt.wantAuth && w.Header().Get("Authorization") == "" {
				t.Error("expected Authorization header, got empty")
			}
			if !tt.wantAuth && w.Header().Get("Authorization") != "" {
				t.Error("did not expect Authorization header, but got one")
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	initTestJWT(t)
	logger := zap.NewNop().Sugar()
	hashed, _ := utils.HashPassword("password")

	tests := []struct {
		name        string
		body        string
		mockGetUser func(ctx context.Context, login string) (*model.User, error)
		wantCode    int
		wantAuth    bool
	}{
		{
			name: "success",
			body: `{"login":"testuser","password":"password"}`,
			mockGetUser: func(ctx context.Context, login string) (*model.User, error) {
				return &model.User{ID: 1, Login: login, PasswordHash: hashed}, nil
			},
			wantCode: http.StatusOK,
			wantAuth: true,
		},
		{
			name: "invalid password",
			body: `{"login":"testuser","password":"wrong"}`,
			mockGetUser: func(ctx context.Context, login string) (*model.User, error) {
				return &model.User{ID: 1, Login: login, PasswordHash: hashed}, nil
			},
			wantCode: http.StatusUnauthorized,
			wantAuth: false,
		},
		{
			name: "user not found",
			body: `{"login":"unknown","password":"pass"}`,
			mockGetUser: func(ctx context.Context, login string) (*model.User, error) {
				return nil, repository.ErrUserNotFound
			},
			wantCode: http.StatusUnauthorized,
			wantAuth: false,
		},
		{
			name:        "invalid json",
			body:        `{"login":"test"`,
			mockGetUser: nil,
			wantCode:    http.StatusBadRequest,
			wantAuth:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{getUserByLoginFunc: tt.mockGetUser}
			svc := service.NewService(repo, "", logger)
			h := NewHandler(svc, logger)

			router := chi.NewRouter()
			router.Post("/api/user/login", h.Login)

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
			if tt.wantAuth && w.Header().Get("Authorization") == "" {
				t.Error("expected Authorization header, got empty")
			}
		})
	}
}

func TestHandler_UploadOrder(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name       string
		body       string
		userID     int
		mockInsert func(ctx context.Context, number string, userID int) (bool, error)
		wantCode   int
	}{
		{
			name:   "new order accepted",
			body:   "4532015112830366", // валидный по Луна
			userID: 1,
			mockInsert: func(ctx context.Context, number string, userID int) (bool, error) {
				return true, nil
			},
			wantCode: http.StatusAccepted,
		},
		{
			name:   "already uploaded by this user",
			body:   "4532015112830366",
			userID: 1,
			mockInsert: func(ctx context.Context, number string, userID int) (bool, error) {
				return false, nil
			},
			wantCode: http.StatusOK,
		},
		{
			name:       "invalid order number (Luhn)",
			body:       "123456",
			userID:     1,
			mockInsert: nil,
			wantCode:   http.StatusUnprocessableEntity,
		},
		{
			name:   "order already exists by another user",
			body:   "4532015112830366",
			userID: 1,
			mockInsert: func(ctx context.Context, number string, userID int) (bool, error) {
				return false, repository.ErrOrderAlreadyExists
			},
			wantCode: http.StatusConflict,
		},
		{
			name:       "empty body",
			body:       "",
			userID:     1,
			mockInsert: nil,
			wantCode:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{insertOrderFunc: tt.mockInsert}
			svc := service.NewService(repo, "", logger)
			h := NewHandler(svc, logger)

			router := chi.NewRouter()
			router.Post("/api/user/orders", h.UploadOrder)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(tt.body)))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestHandler_GetOrders(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name       string
		userID     int
		mockOrders func(ctx context.Context, userID int) ([]model.Order, error)
		wantCode   int
		wantLen    int
	}{
		{
			name:   "with orders",
			userID: 1,
			mockOrders: func(ctx context.Context, userID int) ([]model.Order, error) {
				return []model.Order{
					{Number: "123", Status: model.OrderStatusProcessed, Accrual: ptr(100), UploadedAt: time.Now()},
				}, nil
			},
			wantCode: http.StatusOK,
			wantLen:  1,
		},
		{
			name:   "no orders",
			userID: 1,
			mockOrders: func(ctx context.Context, userID int) ([]model.Order, error) {
				return []model.Order{}, nil
			},
			wantCode: http.StatusNoContent,
			wantLen:  0,
		},
		{
			name:   "db error",
			userID: 1,
			mockOrders: func(ctx context.Context, userID int) ([]model.Order, error) {
				return nil, errors.New("db error")
			},
			wantCode: http.StatusInternalServerError,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{getUserOrdersFunc: tt.mockOrders}
			svc := service.NewService(repo, "", logger)
			h := NewHandler(svc, logger)

			router := chi.NewRouter()
			router.Get("/api/user/orders", h.GetOrders)

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
			if tt.wantLen > 0 {
				var orders []model.OrderResponse
				if err := json.NewDecoder(w.Body).Decode(&orders); err != nil {
					t.Fatal(err)
				}
				if len(orders) != tt.wantLen {
					t.Errorf("got %d orders, want %d", len(orders), tt.wantLen)
				}
			}
		})
	}
}

func TestHandler_GetBalance(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name          string
		userID        int
		mockBalance   func(ctx context.Context, userID int) (float64, float64, error)
		wantCode      int
		wantCurrent   float64
		wantWithdrawn float64
	}{
		{
			name:   "success",
			userID: 1,
			mockBalance: func(ctx context.Context, userID int) (float64, float64, error) {
				return 150.5, 20.0, nil
			},
			wantCode:      http.StatusOK,
			wantCurrent:   150.5,
			wantWithdrawn: 20.0,
		},
		{
			name:   "db error",
			userID: 1,
			mockBalance: func(ctx context.Context, userID int) (float64, float64, error) {
				return 0, 0, errors.New("db error")
			},
			wantCode:      http.StatusInternalServerError,
			wantCurrent:   0,
			wantWithdrawn: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{getUserBalanceFunc: tt.mockBalance}
			svc := service.NewService(repo, "", logger)
			h := NewHandler(svc, logger)

			router := chi.NewRouter()
			router.Get("/api/user/balance", h.GetBalance)

			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
			if tt.wantCode == http.StatusOK {
				var resp model.BalanceResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatal(err)
				}
				if resp.Current != tt.wantCurrent || resp.Withdrawn != tt.wantWithdrawn {
					t.Errorf("got %+v, want Current=%f, Withdrawn=%f", resp, tt.wantCurrent, tt.wantWithdrawn)
				}
			}
		})
	}
}

func TestHandler_Withdraw(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name         string
		userID       int
		body         string
		mockWithdraw func(ctx context.Context, userID int, orderNumber string, sum float64) error
		wantCode     int
	}{
		{
			name:   "success",
			userID: 1,
			body:   `{"order":"4532015112830366","sum":50}`,
			mockWithdraw: func(ctx context.Context, userID int, orderNumber string, sum float64) error {
				return nil
			},
			wantCode: http.StatusOK,
		},
		{
			name:         "invalid order number",
			userID:       1,
			body:         `{"order":"123","sum":10}`,
			mockWithdraw: nil,
			wantCode:     http.StatusUnprocessableEntity,
		},
		{
			name:   "insufficient funds",
			userID: 1,
			body:   `{"order":"4532015112830366","sum":100}`,
			mockWithdraw: func(ctx context.Context, userID int, orderNumber string, sum float64) error {
				return repository.ErrInsufficientFunds
			},
			wantCode: http.StatusPaymentRequired,
		},
		{
			name:         "invalid json",
			userID:       1,
			body:         `{"order":"123"`,
			mockWithdraw: nil,
			wantCode:     http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{withdrawFunc: tt.mockWithdraw}
			svc := service.NewService(repo, "", logger)
			h := NewHandler(svc, logger)

			router := chi.NewRouter()
			router.Post("/api/user/balance/withdraw", h.Withdraw)

			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestHandler_GetWithdrawals(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name          string
		userID        int
		mockWithdraws func(ctx context.Context, userID int) ([]model.Withdrawal, error)
		wantCode      int
		wantLen       int
	}{
		{
			name:   "with withdrawals",
			userID: 1,
			mockWithdraws: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
				return []model.Withdrawal{
					{OrderNumber: "123", Sum: 50, ProcessedAt: time.Now()},
				}, nil
			},
			wantCode: http.StatusOK,
			wantLen:  1,
		},
		{
			name:   "no withdrawals",
			userID: 1,
			mockWithdraws: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
				return []model.Withdrawal{}, nil
			},
			wantCode: http.StatusNoContent,
			wantLen:  0,
		},
		{
			name:   "db error",
			userID: 1,
			mockWithdraws: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
				return nil, errors.New("db error")
			},
			wantCode: http.StatusInternalServerError,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{getUserWithdrawalsFunc: tt.mockWithdraws}
			svc := service.NewService(repo, "", logger)
			h := NewHandler(svc, logger)

			router := chi.NewRouter()
			router.Get("/api/user/withdrawals", h.GetWithdrawals)

			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
			if tt.wantLen > 0 {
				var withdrawals []model.WithdrawalResponse
				if err := json.NewDecoder(w.Body).Decode(&withdrawals); err != nil {
					t.Fatal(err)
				}
				if len(withdrawals) != tt.wantLen {
					t.Errorf("got %d withdrawals, want %d", len(withdrawals), tt.wantLen)
				}
			}
		})
	}
}
