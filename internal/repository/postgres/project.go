package postgres

import (
	"context"
	"database/sql"
	"github.com/AleksKAG/ai-construction-manager/internal/domain"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ProjectRepository) FindByID(ctx context.Context, id uint) (*domain.Project, error) {
	var p domain.Project
	err := r.db.WithContext(ctx).
		Preload("Tasks").
		Preload("Tasks.Resources").
		Preload("Documents").
		First(&p, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, sql.ErrNoRows
	}
	return &p, err
}

func (r *ProjectRepository) Update(ctx context.Context, p *domain.Project) error {
	return r.db.WithContext(ctx).Save(p).Error
}
