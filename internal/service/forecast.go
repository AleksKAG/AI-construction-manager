package service

import (
	"stroy-assistent/internal/domain"
	"stroy-assistent/pkg/gonum"
	"time"
)

type ForecastResult struct {
	EstimatedCompletionDate time.Time
	DelayRisk               string // low, medium, high
	DaysBehindSchedule      int
	Confidence              float64 // 0..1
}

// ForecastCompletion прогнозирует дату завершения на основе фактической выработки
func ForecastCompletion(project domain.Project, actualProgress []struct {
	Day     int     // день от начала проекта
	Progress float64 // 0.0 - 1.0
}) ForecastResult {
	// Преобразуем данные для регрессии
	var days, progress []float64
	for _, p := range actualProgress {
		days = append(days, float64(p.Day))
		progress = append(progress, p.Progress)
	}

	lr := &gonum.LinearRegression{}
	lr.Fit(days, progress)

	// Прогнозируем дни до завершения
	daysToComplete := lr.DaysToCompletion()

	// Определяем риск отставания
	risk := "low"
	if daysToComplete > float64(project.DurationDays())*1.3 {
		risk = "high"
	} else if daysToComplete > float64(project.DurationDays())*1.15 {
		risk = "medium"
	}

	// Рассчитываем дату завершения
	startDate := time.Now()
	if project.StartDate != nil {
		startDate = *project.StartDate
	}
	completionDate := startDate.AddDate(0, 0, int(daysToComplete))

	return ForecastResult{
		EstimatedCompletionDate: completionDate,
		DelayRisk:               risk,
		DaysBehindSchedule:      int(daysToComplete) - project.DurationDays(),
		Confidence:              calculateConfidence(len(actualProgress)),
	}
}

func (p *domain.Project) DurationDays() int {
	total := 0
	for _, t := range p.Tasks {
		total += t.DurationDays
	}
	return total
}

func calculateConfidence(samples int) float64 {
	if samples < 3 {
		return 0.3
	} else if samples < 7 {
		return 0.6
	}
	return 0.85
}
