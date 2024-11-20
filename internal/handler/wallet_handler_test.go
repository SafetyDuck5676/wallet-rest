package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"wallet/internal/handler"
	"wallet/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock für WalletService
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error {
	args := m.Called(ctx, walletID, amount)
	return args.Error(0)
}

func (m *MockWalletService) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	args := m.Called(ctx, walletID, amount)
	return args.Error(0)
}

func (m *MockWalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(int64), args.Error(1)
}

// Test für HandleWalletOperation
func TestHandleWalletOperation(t *testing.T) {
	mockService := new(MockWalletService)
	handler := handler.NewWalletHandler(mockService)

	t.Run("successful deposit", func(t *testing.T) {
		walletID := uuid.New()
		reqBody := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "DEPOSIT",
			"amount":        100,
		}

		mockService.On("Deposit", mock.Anything, walletID, int64(100)).Return(nil)

		reqBodyJSON, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/wallet/operation", bytes.NewReader(reqBodyJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.HandleWalletOperation(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid operation type", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"walletId":      uuid.New().String(),
			"operationType": "INVALID",
			"amount":        100,
		}

		reqBodyJSON, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/wallet/operation", bytes.NewReader(reqBodyJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.HandleWalletOperation(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("withdraw fails with insufficient balance", func(t *testing.T) {
		walletID := uuid.New()
		reqBody := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "WITHDRAW",
			"amount":        100,
		}

		mockService.On("Withdraw", mock.Anything, walletID, int64(100)).Return(service.ErrInsufficientFunds)

		reqBodyJSON, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/wallet/operation", bytes.NewReader(reqBodyJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.HandleWalletOperation(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})
}

// Test für GetBalance
func TestGetBalance(t *testing.T) {
	mockService := new(MockWalletService)
	handler := handler.NewWalletHandler(mockService)

	t.Run("successful balance retrieval", func(t *testing.T) {
		walletID := uuid.New()
		mockService.On("GetBalance", mock.Anything, walletID).Return(int64(500), nil)

		req := httptest.NewRequest(http.MethodGet, "/wallet/"+walletID.String(), nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/wallet/{id}", handler.GetBalance)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var response map[string]int64
		err := json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)
		require.Equal(t, int64(500), response["balance"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid wallet ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/wallet/invalid-id", nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/wallet/{id}", handler.GetBalance)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("balance retrieval fails", func(t *testing.T) {
		walletID := uuid.New()
		mockService.On("GetBalance", mock.Anything, walletID).Return(int64(0), service.ErrWalletNotFound)

		req := httptest.NewRequest(http.MethodGet, "/wallet/"+walletID.String(), nil)
		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/wallet/{id}", handler.GetBalance)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})
}
