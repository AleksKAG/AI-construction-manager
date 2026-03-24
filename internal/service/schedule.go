package service

import (
	"errors"
	"sort"
	"github.com/AleksKAG/ai-construction-manager/internal/domain"
	"time"
)

// BuildSchedule строит Gantt-график с учётом зависимостей (алгоритм Кана)
func BuildSchedule(tasks []domain.Task, projectStartDate time.Time) ([]domain.Task, error) {
	// Карта зависимостей: задача -> количество незавершённых предшественников
	inDegree := make(map[uint]int)
	graph := make(map[uint][]uint) // предшественник -> потомки

	// Инициализация
	taskMap := make(map[uint]*domain.Task)
	for i := range tasks {
		task := &tasks[i]
		taskMap[task.ID] = task
		inDegree[task.ID] = len(task.Dependencies)

		for _, depID := range task.Dependencies {
			graph[depID] = append(graph[depID], task.ID)
		}
	}

	// Очередь задач без зависимостей
	queue := []uint{}
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	// Топологическая сортировка
	sorted := []domain.Task{}
	currentDate := projectStartDate

	for len(queue) > 0 {
		// Сортируем для детерминированности
		sort.Slice(queue, func(i, j int) bool { return queue[i] < queue[j] })

		id := queue[0]
		queue = queue[1:]

		task := *taskMap[id]
		task.StartDate = &currentDate
		endDate := currentDate.AddDate(0, 0, task.DurationDays)
		task.EndDate = &endDate
		task.Status = "scheduled"
		sorted = append(sorted, task)

		// Обновляем зависимости потомков
		for _, childID := range graph[id] {
			inDegree[childID]--
			if inDegree[childID] == 0 {
				queue = append(queue, childID)
			}
		}

		// Следующая задача начинается после текущей (упрощённо)
		// В реальности нужно учитывать параллелизм — здесь для простоты
		currentDate = endDate
	}

	// Проверка циклов
	if len(sorted) != len(tasks) {
		return nil, errors.New("cycle detected in task dependencies")
	}

	return sorted, nil
}
