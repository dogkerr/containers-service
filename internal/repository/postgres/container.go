package postgres

import "gorm.io/gorm"


type ContainerRepository struct {
	DB *gorm.DB
}


func NewContainerRepo(db *gorm.DB) *ContainerRepository {
	return &ContainerRepository{db}
}
