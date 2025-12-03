package model

import "time"

type Node struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Hostname  string    `gorm:"size:256;not null" json:"hostname"`
	IP        string    `gorm:"size:64" json:"ip"`
	CPUTotal  int       `gorm:"not null" json:"cpu_total"`
	CPUUsed   int       `gorm:"not null" json:"cpu_used"`
	MemTotal  int       `gorm:"not null" json:"mem_total"`
	MenUsed   int       `gorm:"not null" json:"mem_used"`
	DiskTotal int       `gorm:"not null" json:"disk_total"`
	DiskUsed  int       `gorm:"not null" json:"disk_used"`
	Status    string    `gorm:"size:32;not null" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
