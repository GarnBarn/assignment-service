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
	CreateAssignment(assignment *globalmodel.Assignment) (model.AssignmentPublic, error)
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

func (a *assignmentService) CreateAssignment(assignmentData *globalmodel.Assignment) (result model.AssignmentPublic, err error) {
	ctx := context.Background()
	// Check if requested tag is exists
	response, err := a.tagClient.IsTagExists(ctx, &proto.TagRequest{
		TagId:             int32(assignmentData.TagID),
		ConsealPrivateKey: true,
	})

	if err != nil {
		logrus.Error(err)
		return result, err
	}

	if !response.IsExists {
		logrus.Warn("Inputted tag is not found, ", assignmentData.TagID)
		return result, ErrTagNotFound
	}

	err = a.assignmentRepository.CreateAssignment(assignmentData)
	if err != nil {
		logrus.Error(err)
		return result, err
	}

	// Get Tag Data
	tagResult, err := a.tagClient.GetTag(ctx, &proto.TagRequest{
		TagId:             int32(assignmentData.TagID),
		ConsealPrivateKey: false,
	})
	j, err := json.MarshalIndent(tagResult, "", "\t")

	result = model.ToAssignmentPublic(*assignmentData, (*json.RawMessage)(&j))

	return result, err
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
