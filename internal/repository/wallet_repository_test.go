package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"wallet/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetBalance_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)

	walletID := uuid.New()
	expectedBalance := int64(5000)

	mock.ExpectQuery("SELECT balance FROM wallets WHERE id = \\$1").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(expectedBalance))

	balance, err := repo.GetBalance(context.Background(), walletID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
	mock.ExpectationsWereMet()
}

func TestGetBalance_WalletNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)

	walletID := uuid.New()

	mock.ExpectQuery("SELECT balance FROM wallets WHERE id = \\$1").
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)

	balance, err := repo.GetBalance(context.Background(), walletID)

	assert.Error(t, err)
	assert.Equal(t, int64(0), balance)
	assert.Equal(t, "wallet not found", err.Error())
	mock.ExpectationsWereMet()
}

func TestUpdateBalance_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)

	walletID := uuid.New()
	initialBalance := int64(5000)
	amountToAdd := int64(2000)
	newBalance := initialBalance + amountToAdd

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT balance FROM wallets WHERE id = \\$1 FOR UPDATE").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(initialBalance))
	mock.ExpectExec("UPDATE wallets SET balance = \\$1, updated_at = \\$2 WHERE id = \\$3").
		WithArgs(newBalance, sqlmock.AnyArg(), walletID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.UpdateBalance(context.Background(), walletID, amountToAdd)

	assert.NoError(t, err)
	mock.ExpectationsWereMet()
}

func TestUpdateBalance_WalletNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)

	walletID := uuid.New()
	amountToAdd := int64(2000)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT balance FROM wallets WHERE id = \\$1 FOR UPDATE").
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = repo.UpdateBalance(context.Background(), walletID, amountToAdd)

	assert.Error(t, err)
	assert.Equal(t, "wallet not found", err.Error())
	mock.ExpectationsWereMet()
}

func TestUpdateBalance_InsufficientFunds(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)

	walletID := uuid.New()
	initialBalance := int64(5000)
	amountToWithdraw := int64(-6000)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT balance FROM wallets WHERE id = \\$1 FOR UPDATE").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(initialBalance))
	mock.ExpectRollback()

	err = repo.UpdateBalance(context.Background(), walletID, amountToWithdraw)

	assert.Error(t, err)
	assert.Equal(t, "insufficient funds", err.Error())
	mock.ExpectationsWereMet()
}

func TestGetWallet_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)

	walletID := uuid.New()
	expectedWallet := &model.Wallet{
		ID:        walletID,
		Balance:   5000,
		UpdatedAt: time.Now(),
	}

	mock.ExpectQuery("SELECT id, balance, updated_at FROM wallets WHERE id = \\$1").
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "balance", "updated_at"}).
			AddRow(expectedWallet.ID, expectedWallet.Balance, expectedWallet.UpdatedAt))

	wallet, err := repo.GetWallet(context.Background(), walletID)

	assert.NoError(t, err)
	assert.Equal(t, expectedWallet.ID, wallet.ID)
	assert.Equal(t, expectedWallet.Balance, wallet.Balance)
	mock.ExpectationsWereMet()
}

func TestGetWallet_WalletNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)

	walletID := uuid.New()

	mock.ExpectQuery("SELECT id, balance, updated_at FROM wallets WHERE id = \\$1").
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)

	wallet, err := repo.GetWallet(context.Background(), walletID)

	assert.Error(t, err)
	assert.Nil(t, wallet)
	assert.Equal(t, "wallet not found", err.Error())
	mock.ExpectationsWereMet()
}
