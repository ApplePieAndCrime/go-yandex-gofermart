package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/auth"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/utils"
	"go.uber.org/zap"
)

func initTestJWT(t *testing.T) *auth.JWTManager {
	t.Helper()

	jwtManager, err := auth.NewJWTManager("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	return jwtManager
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

func TestService_RegisterUser_Success(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		createUserFunc: func(ctx context.Context, login, hashed string) (int, error) {
			return 1, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	userID, token, err := s.RegisterUser(context.Background(), "testuser", "password")
	if err != nil {
		t.Fatalf("RegisterUser error: %v", err)
	}
	if userID != 1 {
		t.Errorf("userID = %d, want 1", userID)
	}
	if token == "" {
		t.Error("token is empty")
	}
}

func TestService_RegisterUser_DuplicateLogin(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		createUserFunc: func(ctx context.Context, login, hashed string) (int, error) {
			return 0, repository.ErrLoginAlreadyExists
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	_, _, err := s.RegisterUser(context.Background(), "testuser", "password")
	if !errors.Is(err, repository.ErrLoginAlreadyExists) {
		t.Errorf("expected ErrLoginAlreadyExists, got %v", err)
	}
}

func TestService_LoginUser_Success(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	hashed, _ := utils.HashPassword("password")
	repo := &mockRepo{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return &model.User{ID: 1, Login: login, PasswordHash: hashed}, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	userID, token, err := s.LoginUser(context.Background(), "testuser", "password")
	if err != nil {
		t.Fatalf("LoginUser error: %v", err)
	}
	if userID != 1 {
		t.Errorf("userID = %d, want 1", userID)
	}
	if token == "" {
		t.Error("token is empty")
	}
}

func TestService_LoginUser_InvalidPassword(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	hashed, _ := utils.HashPassword("password")
	repo := &mockRepo{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return &model.User{ID: 1, Login: login, PasswordHash: hashed}, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	_, _, err := s.LoginUser(context.Background(), "testuser", "wrong")
	if err == nil || err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials', got %v", err)
	}
}

func TestService_LoginUser_UserNotFound(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		getUserByLoginFunc: func(ctx context.Context, login string) (*model.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	_, _, err := s.LoginUser(context.Background(), "testuser", "password")
	if err == nil || err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials', got %v", err)
	}
}

func TestService_UploadOrder_Success(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		insertOrderFunc: func(ctx context.Context, number string, userID int) (bool, error) {
			return true, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	status, err := s.UploadOrder(context.Background(), 1, "4532015112830366")
	if err != nil {
		t.Fatal(err)
	}
	if status != OrderNewAccepted {
		t.Errorf("status = %s, want %s", status, OrderNewAccepted)
	}
}

func TestService_UploadOrder_AlreadyUploaded(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		insertOrderFunc: func(ctx context.Context, number string, userID int) (bool, error) {
			return false, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	status, err := s.UploadOrder(context.Background(), 1, "4532015112830366")
	if err != nil {
		t.Fatal(err)
	}
	if status != OrderAlreadyUploaded {
		t.Errorf("status = %s, want %s", status, OrderAlreadyUploaded)
	}
}

func TestService_UploadOrder_InvalidNumber(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{}
	s := NewService(repo, "", logger, jwtManager)

	_, err := s.UploadOrder(context.Background(), 1, "123456")
	if !errors.Is(err, ErrInvalidOrderNumber) {
		t.Errorf("expected ErrInvalidOrderNumber, got %v", err)
	}
}

func TestService_UploadOrder_AlreadyExistsByOther(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		insertOrderFunc: func(ctx context.Context, number string, userID int) (bool, error) {
			return false, repository.ErrOrderAlreadyExists
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	_, err := s.UploadOrder(context.Background(), 2, "4532015112830366")
	if !errors.Is(err, ErrOrderAlreadyExists) {
		t.Errorf("expected ErrOrderAlreadyExists, got %v", err)
	}
}

func TestService_GetUserOrders_Success(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	expectedOrders := []model.Order{
		{Number: "123", Status: "PROCESSED", Accrual: ptr(100.0)},
	}
	repo := &mockRepo{
		getUserOrdersFunc: func(ctx context.Context, userID int) ([]model.Order, error) {
			return expectedOrders, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	orders, err := s.GetUserOrders(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserOrders error: %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("len = %d, want 1", len(orders))
	}
	if orders[0].Number != "123" {
		t.Errorf("number = %s, want 123", orders[0].Number)
	}
}

func TestService_GetUserOrders_Empty(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		getUserOrdersFunc: func(ctx context.Context, userID int) ([]model.Order, error) {
			return []model.Order{}, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	orders, err := s.GetUserOrders(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserOrders error: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("len = %d, want 0", len(orders))
	}
}

func TestService_GetBalance_Success(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		getUserBalanceFunc: func(ctx context.Context, userID int) (float64, float64, error) {
			return 150.5, 20.0, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	balance, err := s.GetBalance(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetBalance error: %v", err)
	}
	if balance.Current != 150.5 || balance.Withdrawn != 20.0 {
		t.Errorf("balance = %+v, want {Current:150.5, Withdrawn:20.0}", balance)
	}
}

func TestService_GetBalance_Error(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		getUserBalanceFunc: func(ctx context.Context, userID int) (float64, float64, error) {
			return 0, 0, errors.New("db error")
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	_, err := s.GetBalance(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestService_Withdraw_Success(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		withdrawFunc: func(ctx context.Context, userID int, orderNumber string, sum float64) error {
			return nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	err := s.Withdraw(context.Background(), 1, "4532015112830366", 50.0)
	if err != nil {
		t.Fatalf("Withdraw error: %v", err)
	}
}

func TestService_Withdraw_InvalidOrderNumber(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{}
	s := NewService(repo, "", logger, jwtManager)

	err := s.Withdraw(context.Background(), 1, "123", 10)
	if !errors.Is(err, ErrInvalidOrderNumber) {
		t.Errorf("expected ErrInvalidOrderNumber, got %v", err)
	}
}

func TestService_Withdraw_InsufficientFunds(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		withdrawFunc: func(ctx context.Context, userID int, orderNumber string, sum float64) error {
			return repository.ErrInsufficientFunds
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	err := s.Withdraw(context.Background(), 1, "4532015112830366", 100)
	if !errors.Is(err, repository.ErrInsufficientFunds) {
		t.Errorf("expected repository.ErrInsufficientFunds, got %v", err)
	}
}

func TestService_GetWithdrawals_Success(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	expected := []model.Withdrawal{
		{OrderNumber: "123", Sum: 50.0},
		{OrderNumber: "456", Sum: 30.0},
	}
	repo := &mockRepo{
		getUserWithdrawalsFunc: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
			return expected, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	withdrawals, err := s.GetWithdrawals(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetWithdrawals error: %v", err)
	}
	if len(withdrawals) != 2 {
		t.Errorf("len = %d, want 2", len(withdrawals))
	}
	if withdrawals[0].Order != "123" {
		t.Errorf("first order = %s, want 123", withdrawals[0].Order)
	}
}

func TestService_GetWithdrawals_Empty(t *testing.T) {
	jwtManager := initTestJWT(t)
	logger := zap.NewNop().Sugar()
	repo := &mockRepo{
		getUserWithdrawalsFunc: func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
			return []model.Withdrawal{}, nil
		},
	}
	s := NewService(repo, "", logger, jwtManager)

	withdrawals, err := s.GetWithdrawals(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetWithdrawals error: %v", err)
	}
	if len(withdrawals) != 0 {
		t.Errorf("len = %d, want 0", len(withdrawals))
	}
}

func ptr(f float64) *float64 {
	return &f
}
