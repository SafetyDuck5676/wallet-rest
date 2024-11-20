package model

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	Balance   int64     `json:"balance"`
	UpdatedAt time.Time `json:"updatedAt"`
}
