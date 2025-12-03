package model

import (
	"time"
)

type VM struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:64;not null" json:"name"`
	CPU          int       `gorm:"not null" json:"CPU"`
	MemoryMB     int       `gorm:"not null" json:"memory_mb"`
	DiskGB       int       `gorm:"not null" json:"disk_gb"`
	Status       string    `gorm:"size:32;not null" json:"status"`
	Description  string    `gorm:"size:255;not null" json:"description"`
	HypervisorID string    `gorm:"size:64;uniqueIndex;not null" json:"hypervisor_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
