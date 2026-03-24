package domain

import (
	"time"
)

type Project struct {
	ID          uint       `gorm:"primaryKey"`
	Name        string     `gorm:"not null"`
	Address     string
	StartDate   *time.Time
	EndDate     *time.Time
	Budget      float64    `gorm:"type:numeric(15,2)"`
	Status      string     `gorm:"default:'planning'"` // planning, active, completed, delayed
	Tasks       []Task
	Documents   []Document
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

type Task struct {
	ID            uint       `gorm:"primaryKey"`
	ProjectID     uint       `gorm:"index"`
	Name          string     `gorm:"not null"`
	Description   string
	DurationDays  int        `gorm:"not null"` // плановая длительность в днях
	StartDate     *time.Time
	EndDate       *time.Time
	Dependencies  []uint     `gorm:"type:jsonb;default:'[]'"` // ID задач-предшественников
	ActualProgress float64   `gorm:"default:0"` // 0.0 - 1.0 (фактическая выработка)
	Resources     []Resource
	Status        string     `gorm:"default:'pending'"` // pending, in_progress, done, blocked
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
}

type Resource struct {
	ID          uint    `gorm:"primaryKey"`
	TaskID      uint    `gorm:"index"`
	Type        string  `gorm:"not null"` // labor, material, equipment
	Name        string  `gorm:"not null"`
	Quantity    float64 `gorm:"not null"`
	Unit        string  `gorm:"not null"` // чел, м3, шт, час
	UnitCost    float64 `gorm:"type:numeric(10,2);not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type Document struct {
	ID          uint   `gorm:"primaryKey"`
	ProjectID   uint   `gorm:"index"`
	Name        string `gorm:"not null"`
	Path        string `gorm:"not null"` // путь к файлу на сервере/облаке
	Hash        string `gorm:"not null"` // SHA256 хеш для отслеживания изменений
	Stage       string `gorm:"not null"` // P (проектная), R (рабочая)
	DocType     string `gorm:"not null"` // II (исходные данные), other
	Version     string `gorm:"not null"` // 1.0, 1.1, 2.0
	DurationDays int `json:"duration_days" gorm:"default:0"`
	ChangedAt   time.Time
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
