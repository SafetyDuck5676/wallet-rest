package model

import (
    "github.com/google/uuid"
    "time"
)

type Wallet struct {
    ID        uuid.UUID json:"id"
    Balance   int64     json:"balance"
    UpdatedAt time.Time json:"updatedAt"
}