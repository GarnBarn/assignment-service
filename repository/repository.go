package repository

import (
	"time"

	"github.com/GarnBarn/common-go/model"
	"gorm.io/gorm"
)

type AssignmentRepository interface {
	GetAllAssignment(formPresent bool) ([]model.Assignment, error)
	GetByID(id int) (*model.Assignment, error)
	Update(assignment *model.Assignment) error
}

type assignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	// Migrate the db
	db.AutoMigrate(&model.Assignment{})

	return &assignmentRepository{
		db: db,
	}
}

func (a *assignmentRepository) GetAllAssignment(fromPresent bool) (result []model.Assignment, err error) {
	now := time.Now()

	baseQuery := a.db.Model(&model.Assignment{})

	var res *gorm.DB
	if fromPresent {
		res = baseQuery.Where("due_date >= ?", now.Unix()*1000).Find(&result)
	} else {
		res = baseQuery.Find(&result)
	}

	if res.Error != nil {
		return result, res.Error
	}

	return result, nil
}

func (a *assignmentRepository) GetByID(id int) (*model.Assignment, error) {
	var result model.Assignment
	response := a.db.First(&result, id)
	return &result, response.Error
}

func (a *assignmentRepository) Update(assignment *model.Assignment) error {
	result := a.db.Save(assignment)
	return result.Error
}
