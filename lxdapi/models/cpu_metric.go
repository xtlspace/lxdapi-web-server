package models

import "time"

type CPUMetric struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"index:idx_cpu_name_time,priority:1;size:255" json:"name"`
	CPUUsage  float64   `gorm:"type:decimal(5,2)" json:"cpu_usage"`
	CreatedAt time.Time `gorm:"index:idx_cpu_name_time,priority:2" json:"created_at"`
}
