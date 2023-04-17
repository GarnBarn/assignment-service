package handler

import (
	"errors"
	"net/http"

	"github.com/GarnBarn/common-go/httpserver"
	"github.com/GarnBarn/gb-assignment-service/model"
	"github.com/GarnBarn/gb-assignment-service/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AssignmentHandler struct {
	validate          validator.Validate
	assignmentService service.AssignmentService
}

func NewAssignmentHandler(validate validator.Validate, assignmentService service.AssignmentService) AssignmentHandler {
	return AssignmentHandler{
		validate:          validate,
		assignmentService: assignmentService,
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

	if err := a.assignmentService.CreateAssignment(&assignment); err != nil {
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

	// TODO: Fill tag result.
	assignmentPublic := model.ToAssignmentPublic(assignment, nil)

	c.JSON(http.StatusCreated, assignmentPublic)

}
