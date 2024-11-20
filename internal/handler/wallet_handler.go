package handler

import (
	"encoding/json"
	"net/http"
	"wallet/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type WalletHandler struct {
	service service.WalletServiceInterface
}

func NewWalletHandler(service service.WalletServiceInterface) *WalletHandler {
	return &WalletHandler{service: service}
}

type WalletOperationRequest struct {
	WalletID      uuid.UUID `json:"walletId" `
	OperationType string    `json:"operationType" `
	Amount        int64     `json:"amount"`
}

func (h *WalletHandler) HandleWalletOperation(w http.ResponseWriter, r *http.Request) {
	var req WalletOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	if req.OperationType == "DEPOSIT" {
		err = h.service.Deposit(r.Context(), req.WalletID, req.Amount)
	} else if req.OperationType == "WITHDRAW" {
		err = h.service.Withdraw(r.Context(), req.WalletID, req.Amount)
	} else {
		http.Error(w, "Invalid operation type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	balance, err := h.service.GetBalance(r.Context(), walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int64{"balance": balance})
}
