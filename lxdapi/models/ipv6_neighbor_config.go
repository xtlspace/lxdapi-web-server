package models

import "time"

type IPv6NeighborConfig struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Enabled   bool      `gorm:"default:false" json:"enabled"`
	Iface     string    `gorm:"size:64;default:''" json:"iface"`
	Prefix    string    `gorm:"size:64;default:''" json:"prefix"`
	Gateway   string    `gorm:"size:64;default:''" json:"gateway"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (IPv6NeighborConfig) TableName() string {
	return "ipv6_neighbor_config"
}
