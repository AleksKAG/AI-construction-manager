package repository

import (
	"context"
	"errors"

	"github.com/AleksKAG/ai-construction-manager/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	GetAllObjects(ctx context.Context) ([]models.ProjectObject, error)
	CreateObject(ctx context.Context, obj *models.ProjectObject) error
	GetObject(ctx context.Context, id string) (*models.ProjectObject, error)
	UpdateObject(ctx context.Context, obj *models.ProjectObject) error
	DeleteObject(ctx context.Context, id string) error
	GetAllGraphs(ctx context.Context) ([]models.ProjectGraph, error)
	GetGraph(ctx context.Context, objectID string) (*models.ProjectGraph, error)
	CreateGraph(ctx context.Context, graph *models.ProjectGraph) error
	// Добавьте для tasks, etc.
}

type projectRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewProjectRepository(db *gorm.DB, redis *redis.Client) ProjectRepository {
	return &projectRepo{db: db, redis: redis}
}

func (r *projectRepo) GetAllObjects(ctx context.Context) ([]models.ProjectObject, error) {
	var objs []models.ProjectObject
	if err := r.db.WithContext(ctx).Find(&objs).Error; err != nil {
		return nil, err
	}
	return objs, nil
}

func (r *projectRepo) CreateObject(ctx context.Context, obj *models.ProjectObject) error {
	return r.db.WithContext(ctx).Create(obj).Error
}

func (r *projectRepo) GetObject(ctx context.Context, id string) (*models.ProjectObject, error) {
	var obj models.ProjectObject
	if err := r.db.WithContext(ctx).First(&obj, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &obj, nil
}

func (r *projectRepo) UpdateObject(ctx context.Context, obj *models.ProjectObject) error {
	return r.db.WithContext(ctx).Save(obj).Error
}

func (r *projectRepo) DeleteObject(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ProjectObject{}, "id = ?", id).Error
}

func (r *projectRepo) GetAllGraphs(ctx context.Context) ([]models.ProjectGraph, error) {
	var graphs []models.ProjectGraph
	// Join с tasks
	if err := r.db.WithContext(ctx).Preload("Tasks").Find(&graphs).Error; err != nil {
		return nil, err
	}
	return graphs, nil
}

func (r *projectRepo) GetGraph(ctx context.Context, objectID string) (*models.ProjectGraph, error) {
	var graph models.ProjectGraph
	if err := r.db.WithContext(ctx).Preload("Tasks").First(&graph, "object_id = ?", objectID).Error; err != nil {
		return nil, err
	}
	return &graph, nil
}

func (r *projectRepo) CreateGraph(ctx context.Context, graph *models.ProjectGraph) error {
	return r.db.WithContext(ctx).Create(graph).Error
}
