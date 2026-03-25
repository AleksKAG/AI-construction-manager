package models

type ProjectObject struct {
	ID            string            `gorm:"primaryKey" json:"id"`
	Name          string            `json:"name"`
	Characteristics map[string]string `gorm:"type:jsonb" json:"characteristics"`
	CostEstimates map[string]float64 `gorm:"type:jsonb" json:"cost_estimates"`
}

type GanttTask struct {
	ID        string `gorm:"primaryKey" json:"id"`
	ObjectID  string `json:"object_id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Duration  int    `json:"duration"`
}

type ProjectGraph struct {
	ID string `gorm:"primaryKey" json:"id"`
	ObjectID string      `json:"object_id"`
	Tasks    []GanttTask `json:"tasks"`
}
