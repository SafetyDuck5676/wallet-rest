package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
	"wallet/internal/model"

	"github.com/google/uuid"
)

type Wallet struct {
	ID      uuid.UUID
	Balance int64
}

type WalletRepository struct {
	db *sql.DB
	mu sync.Mutex // To handle concurrent operations
}

type WalletRepositoryInterface interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) error
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func LoadConfigFromEnv() Config {
	return Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
}

func NewPostgresDB(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection check
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connection pool
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * 60)

	return db, nil
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	var balance int64
	err := r.db.QueryRowContext(ctx, "SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, errors.New("wallet not found")
	}
	return balance, err
}

func (r *WalletRepository) GetWallet(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.QueryRowContext(ctx, "SELECT id, balance, updated_at FROM wallets WHERE id = $1", walletID).
		Scan(&wallet.ID, &wallet.Balance, &wallet.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("wallet not found")
	}
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var balance int64
	err = tx.QueryRow("SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", walletID).Scan(&balance)
	if err == sql.ErrNoRows {
		return errors.New("wallet not found")
	} else if err != nil {
		tx.Rollback()
		return err
	}

	newBalance := balance + amount
	if newBalance < 0 {
		tx.Rollback()
		return errors.New("insufficient funds")
	}

	_, err = tx.Exec("UPDATE wallets SET balance = $1, updated_at = $2 WHERE id = $3", newBalance, time.Now(), walletID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
