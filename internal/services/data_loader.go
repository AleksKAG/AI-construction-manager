package services

import (
	"context"

	"github.com/AleksKAG/ai-construction-manager/internal/models"
	"github.com/AleksKAG/ai-construction-manager/internal/repository"
	"github.com/google/uuid"
)

func LoadSampleData(repo repository.ProjectRepository) error {
	ctx := context.Background()

	// Sample object 1
	obj1 := models.ProjectObject{
		ID:   uuid.New().String(),
		Name: "Клинико-диагностическая лаборатория (КДЛ)",
		Characteristics: map[string]string{
			"Location":    "Белорусская",
			"Type":        "Лаборатория",
			"StartDate":   "20-26 августа",
			"EndDate":     "4 февраля - 5 февраля",
			"Description": "Включает технические отчеты, обмерные чертежи, рабочую документацию и т.д.",
		},
		CostEstimates: map[string]float64{
			"TotalCost":       1000000.0,
			"LaborCost":       400000.0,
			"MaterialCost":    500000.0,
			"EquipmentCost":   100000.0,
			"InflationAdjust": 50000.0,
		},
	}
	if err := repo.CreateObject(ctx, &obj1); err != nil {
		return err
	}

	// Sample graph for obj1
	graph1 := models.ProjectGraph{
		ObjectID: obj1.ID,
		Tasks: []models.GanttTask{
			{ID: uuid.New().String(), ObjectID: obj1.ID, Name: "Технический отчет по сист.отопления", StartDate: "20-26 августа", EndDate: "27 августа - 2 сентября", Duration: 7},
			// ... добавьте больше
		},
	}
	if err := repo.CreateGraph(ctx, &graph1); err != nil {
		return err
	}

	// Аналогично для obj2, obj3
	// ...

	return nil
}
