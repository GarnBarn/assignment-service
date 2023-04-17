package service

import (
	"context"
	"errors"

	"github.com/GarnBarn/common-go/model"
	"github.com/GarnBarn/common-go/proto"
	"github.com/GarnBarn/gb-assignment-service/repository"
	"github.com/sirupsen/logrus"
)

type AssignmentService interface {
	CreateAssignment(assignment *model.Assignment) error
	GetAllAssignment(fromPresent bool) ([]model.Assignment, error)
}

type assignmentService struct {
	tagClient            proto.TagClient
	assignmentRepository repository.AssignmentRepository
}

var (
	ErrTagNotFound = errors.New("tag not found")
)

func NewAssignmentService(tagClient proto.TagClient, assignmentRepository repository.AssignmentRepository) AssignmentService {
	return &assignmentService{
		tagClient:            tagClient,
		assignmentRepository: assignmentRepository,
	}
}

func (a *assignmentService) CreateAssignment(assignmentData *model.Assignment) error {
	ctx := context.Background()
	// Check if requested tag is exists
	response, err := a.tagClient.IsTagExists(ctx, &proto.TagRequest{
		TagId:             int32(assignmentData.TagID),
		ConsealPrivateKey: true,
	})

	if err != nil {
		logrus.Error(err)
		return err
	}

	if !response.IsExists {
		logrus.Warn("Inputted tag is not found, ", assignmentData.TagID)
		return ErrTagNotFound
	}

	return a.assignmentRepository.CreateAssignment(assignmentData)
}

func (a *assignmentService) GetAllAssignment(fromPresent bool) ([]model.Assignment, error) {
	return a.assignmentRepository.GetAllAssignment(fromPresent)
}
