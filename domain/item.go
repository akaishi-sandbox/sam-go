package domain

import (
	"time"
)

type Item struct {
	ItemID         string    `json:"item_id"`
	Gender         string    `json:"gender"`
	Category       string    `json:"category"`
	AccessCounter  int       `json:"access_counter"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}
