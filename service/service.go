package service

import (
	"context"
	"encoding/json"
	"errors"

	globalmodel "github.com/GarnBarn/common-go/model"
	"github.com/GarnBarn/common-go/proto"
	"github.com/GarnBarn/gb-assignment-service/config"
	"github.com/GarnBarn/gb-assignment-service/model"
	"github.com/GarnBarn/gb-assignment-service/repository"
	"github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

type AssignmentService interface {
	CreateAssignment(assignment *globalmodel.Assignment) error
	GetAllAssignment(fromPresent bool) ([]model.AssignmentPublic, error)
	GetAssignmentById(assignmentId int) (model.AssignmentPublic, error)
	UpdateAssignment(updateAssignmentRequest *model.UpdateAssignmentRequest, id int) (model.AssignmentPublic, error)
	DeleteAssignment(assignmentId int) error
}

type assignmentService struct {
	tagClient            proto.TagClient
	assignmentRepository repository.AssignmentRepository
	rabbitmqPublisher    *rabbitmq.Publisher
	appConfig            config.Config
}

var (
	ErrTagNotFound = errors.New("tag not found")
	ErrTagError    = errors.New("tag error")
)

func NewAssignmentService(tagClient proto.TagClient, assignmentRepository repository.AssignmentRepository, rabbitmqPublisher *rabbitmq.Publisher, appConfig config.Config) AssignmentService {
	return &assignmentService{
		tagClient:            tagClient,
		assignmentRepository: assignmentRepository,
		rabbitmqPublisher:    rabbitmqPublisher,
		appConfig:            appConfig,
	}
}

func (a *assignmentService) CreateAssignment(assignmentData *globalmodel.Assignment) (err error) {
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

	assignmentByte, err := json.Marshal(assignmentData)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return a.rabbitmqPublisher.Publish(assignmentByte, []string{"create"}, rabbitmq.WithPublishOptionsExchange(a.appConfig.RABBITMQ_ASSIGNMENT_EXCHANGE))
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

func (a *assignmentService) UpdateAssignment(updateAssignmentRequest *model.UpdateAssignmentRequest, id int) (model.AssignmentPublic, error) {
	assignment, err := a.assignmentRepository.GetByID(id)
	if err != nil {
		logrus.Error(err)
		return model.AssignmentPublic{}, err
	}
	updateAssignmentRequest.UpdateAssignment(assignment)

	// Get Tag from tag service
	ctx := context.Background()
	tagResult, err := a.tagClient.GetTag(ctx, &proto.TagRequest{TagId: int32(assignment.TagID), ConsealPrivateKey: true})
	if err != nil {
		logrus.Warnln("Getting tag error for ", assignment.TagID, " : ", err)
		return model.AssignmentPublic{}, ErrTagError
	}

	// Update Assignment
	err = a.assignmentRepository.Update(assignment)

	j, err := json.MarshalIndent(tagResult, "", "\t")
	if err != nil {
		logrus.Warn("Convert protobuf to json failed:", err)
		return model.AssignmentPublic{}, err
	}

	return model.ToAssignmentPublic(*assignment, (*json.RawMessage)(&j)), err
}

func (a *assignmentService) GetAssignmentById(assignmentId int) (model.AssignmentPublic, error) {
	assignment, err := a.assignmentRepository.GetByID(assignmentId)
	if err != nil {
		logrus.Error(err)
		return model.AssignmentPublic{}, err
	}

	ctx := context.Background()
	tagResult, err := a.tagClient.GetTag(ctx, &proto.TagRequest{TagId: int32(assignment.TagID), ConsealPrivateKey: true})
	if err != nil {
		logrus.Warnln("Getting tag error for ", assignment.TagID, " : ", err)
		return model.AssignmentPublic{}, err
	}

	j, err := json.MarshalIndent(tagResult, "", "\t")
	if err != nil {
		logrus.Warn("Convert protobuf to json failed:", err)
	}

	return model.ToAssignmentPublic(*assignment, (*json.RawMessage)(&j)), nil
}

func (a *assignmentService) DeleteAssignment(assignmentId int) error {
	logrus.Info("Check delete assignment")
	defer logrus.Info("Complete delete assignment")
	assignmentRequestByte, err := json.Marshal(globalmodel.AssignmentDeleteRequest{ID: assignmentId})
	if err != nil {
		logrus.Error(err)
		return err
	}
	return a.rabbitmqPublisher.Publish(assignmentRequestByte, []string{"delete"}, rabbitmq.WithPublishOptionsExchange(a.appConfig.RABBITMQ_ASSIGNMENT_EXCHANGE))
}
