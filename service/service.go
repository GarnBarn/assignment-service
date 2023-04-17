package service

import (
	"context"
	"encoding/json"
	"errors"

	globalmodel "github.com/GarnBarn/common-go/model"
	"github.com/GarnBarn/common-go/proto"
	"github.com/GarnBarn/gb-assignment-service/model"
	"github.com/GarnBarn/gb-assignment-service/repository"
	"github.com/sirupsen/logrus"
)

type AssignmentService interface {
	CreateAssignment(assignment *globalmodel.Assignment) error
	GetAllAssignment(fromPresent bool) ([]model.AssignmentPublic, error)
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

func (a *assignmentService) CreateAssignment(assignmentData *globalmodel.Assignment) error {
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

func (a *assignmentService) GetAllAssignment(fromPresent bool) (result []model.AssignmentPublic, err error) {

	assignments, err := a.assignmentRepository.GetAllAssignment(fromPresent)
	if err != nil {
		return []model.AssignmentPublic{}, err
	}

	// Fillin the tag data into the assignment public model

	ctx := context.Background()
	for _, item := range assignments {
		tagResult, err := a.tagClient.GetTag(ctx, &proto.TagRequest{TagId: int32(item.TagID), ConsealPrivateKey: true})
		if err != nil {
			logrus.Warnln("Getting tag error for ", item.TagID, " : ", err)
			continue
		}

		j, err := json.MarshalIndent(tagResult, "", "\t")
		if err != nil {
			logrus.Warn("Convert protobuf to json failed:", err)
		}

		result = append(result, model.ToAssignmentPublic(item, (*json.RawMessage)(&j)))
	}

	return result, err

}
