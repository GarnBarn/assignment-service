package service

import (
	"github.com/GarnBarn/common-go/model"
	"github.com/GarnBarn/gb-assignment-service/repository"
)

type AssignmentService interface {
	CreateAssignment(assignment *model.Assignment) error
	GetAllAssignment(fromPresent bool) ([]model.Assignment, error)
}

type assignmentService struct {
	assignmentRepository repository.AssignmentRepository
}

func NewAssignmentService(assignmentRepository repository.AssignmentRepository) AssignmentService {
	return &assignmentService{
		assignmentRepository: assignmentRepository,
	}
}

func (a *assignmentService) CreateAssignment(assignmentData *model.Assignment) error {
	return a.assignmentRepository.CreateAssignment(assignmentData)
}

func (a *assignmentService) GetAllAssignment(fromPresent bool) ([]model.Assignment, error) {
	return a.assignmentRepository.GetAllAssignment(fromPresent)
}
