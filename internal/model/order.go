package model

import "time"

type Order struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	VMID        *uint     `gorm:"index" json:"vm_id"`
	AmountCents int       `gorm:"not null" json:"amount_cents"`
	Currency    string    `gorm:"size:8;not null" json:"currency"`
	Status      string    `gorm:"size:32;not null" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at" `
}
