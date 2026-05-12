package auth

import (
	"context"

	"gorm.io/gorm"

	"CallItCureIt/backend/internal/db"
)

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
	GetUserByID(ctx context.Context, id string) (*db.User, error)
	CreateUser(ctx context.Context, user *db.User) error
}

type GormRepository struct {
	database *gorm.DB
}

func NewGormRepository(database *gorm.DB) *GormRepository {
	return &GormRepository{
		database: database,
	}
}

func (r *GormRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*db.User, error) {
	var user db.User

	err := r.database.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormRepository) GetUserByID(
	ctx context.Context,
	id string,
) (*db.User, error) {
	var user db.User

	err := r.database.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormRepository) CreateUser(
	ctx context.Context,
	user *db.User,
) error {
	return r.database.WithContext(ctx).Create(user).Error
}