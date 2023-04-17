package handler

import (
	"context"
	"errors"
	"github.com/GarnBarn/common-go/proto"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/GarnBarn/common-go/httpserver"
	"github.com/GarnBarn/gb-assignment-service/model"
	"github.com/GarnBarn/gb-assignment-service/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AssignmentHandler struct {
	tagClient         proto.TagClient
	validate          validator.Validate
	assignmentService service.AssignmentService
}

var (
	ErrGinBadRequestBody = gin.H{"message": "bad request body."}
)

func NewAssignmentHandler(validate validator.Validate, assignmentService service.AssignmentService, tagClient proto.TagClient) AssignmentHandler {
	return AssignmentHandler{
		validate:          validate,
		assignmentService: assignmentService,
		tagClient:         tagClient,
	}
}

func (a *AssignmentHandler) GetAllAssignment(c *gin.Context) {
	fromPresentString := c.Query("fromPresent")

	logrus.Debug("From Present string: ", fromPresentString)

	fromPresent := true
	if fromPresentString == "" || fromPresentString == "false" {
		fromPresent = false
	}

	logrus.Debug("From Present: ", fromPresent)

	assignments, err := a.assignmentService.GetAllAssignment(fromPresent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.BulkResponse[model.AssignmentPublic]{
		Count:    len(assignments),
		Previous: nil,
		Next:     nil,
		Results:  assignments,
	})
}

func (a *AssignmentHandler) CreateAssignment(c *gin.Context) {
	var assignmentRequest model.AssignmentRequest

	err := c.ShouldBindJSON(&assignmentRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = a.validate.Struct(assignmentRequest)
	if err != nil {
		logrus.Warn("Struct validation failed: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	assignment := assignmentRequest.ToAssignment(c.Param(httpserver.UserUidKey))

	assignmentPublic, err := a.assignmentService.CreateAssignment(&assignment)
	if err != nil {
		if errors.Is(err, service.ErrTagNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, assignmentPublic)

}

func (a *AssignmentHandler) UpdateAssignment(c *gin.Context) {
	assignmentIdString, ok := c.Params.Get("assignmentId")
	if !ok {
		logrus.Warn("Can't get assignmentId from parameters")
		c.JSON(http.StatusBadRequest, ErrGinBadRequestBody)
		return
	}

	// Check if the tagId is int parsable
	assignmentId, err := strconv.Atoi(assignmentIdString)
	if err != nil {
		logrus.Warn("Can't convert assignmentId to int: ", err)
		c.JSON(http.StatusBadRequest, ErrGinBadRequestBody)
		return
	}

	// Bind the request body.
	var updateAssignmentRequest model.UpdateAssignmentRequest
	err = c.ShouldBindJSON(&updateAssignmentRequest)
	if err != nil {
		logrus.Warn("Can't bind request body to model: ", err)
		c.JSON(http.StatusBadRequest, ErrGinBadRequestBody)
		return
	}

	err = a.validate.Struct(updateAssignmentRequest)
	if err != nil {
		logrus.Warn("Struct validation failed: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//Check is tag exist
	updateTagIdRequest := updateAssignmentRequest.TagId
	if updateTagIdRequest != nil {
		ctx := context.Background()
		response, err := a.tagClient.IsTagExists(ctx, &proto.TagRequest{TagId: int32(*updateTagIdRequest)})
		if err != nil {
			logrus.Warnln("Check is tag exist error for ", updateTagIdRequest, " : ", err)
			c.JSON(http.StatusInternalServerError, ErrGinBadRequestBody)
			return
		}
		if response.IsExists == false {
			logrus.Warn("Tag id is not exist")
			c.JSON(http.StatusBadRequest, ErrGinBadRequestBody)
			return
		}
	}

	publicAssignment, err := a.assignmentService.UpdateAssignment(&updateAssignmentRequest, assignmentId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something happen in the server"})
		return
	}

	c.JSON(http.StatusOK, publicAssignment)
}

func (a *AssignmentHandler) GetAssignmentById(c *gin.Context) {
	assignmentIdStr, ok := c.Params.Get("assignmentId")
	if !ok {
		c.JSON(http.StatusBadRequest, ErrGinBadRequestBody)
		return
	}
	assignmentId, err := strconv.Atoi(assignmentIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrGinBadRequestBody)
		return
	}

	publicAssignment, err := a.assignmentService.GetAssignmentById(assignmentId)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "assignment id not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrGinBadRequestBody)
		return
	}
	c.JSON(http.StatusOK, publicAssignment)
}
