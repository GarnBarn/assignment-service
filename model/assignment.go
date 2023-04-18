package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/GarnBarn/common-go/model"
	"github.com/sirupsen/logrus"
)

func ToAssignmentPublic(assignment model.Assignment, tagData *json.RawMessage) AssignmentPublic {
	reminderTime := strings.Split(assignment.ReminderTime, ",")
	reminterTimeInt := []int{}

	for _, item := range reminderTime {
		result, err := strconv.Atoi(item)
		if err != nil {
			logrus.Warn("Can't convert the result to int: ", item, " for ", assignment.ID)
			continue
		}
		reminterTimeInt = append(reminterTimeInt, result)
	}

	assignmentResult := AssignmentPublic{
		ID:           fmt.Sprint(assignment.ID),
		Name:         assignment.Name,
		Author:       assignment.Author,
		Description:  assignment.Description,
		DueDate:      assignment.DueDate,
		Tag:          nil,
		ReminderTime: reminterTimeInt,
	}

	if tagData != nil {
		assignmentResult.Tag = tagData
	}
	return assignmentResult
}

type AssignmentPublic struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Author       string           `json:"author"`
	Description  string           `json:"description,omitempty"`
	DueDate      int              `json:"dueDate"`
	Tag          *json.RawMessage `json:"tag"`
	ReminderTime []int            `json:"reminderTime"`
}

type AssignmentRequest struct {
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description"`
	DueDate      int    `json:"dueDate"`
	TagId        string `json:"tagId"`
	ReminderTime []int  `json:"reminderTime,omitempty" validate:"max=3,omitempty"`
}

func (ar *AssignmentRequest) ToAssignment(author string) model.Assignment {
	reminderTimeByte, _ := json.Marshal(ar.ReminderTime)
	reminderTimeString := strings.Trim(string(reminderTimeByte), "[]")

	tagIdInt, _ := strconv.Atoi(ar.TagId)

	return model.Assignment{
		Name:         ar.Name,
		Author:       author,
		Description:  ar.Description,
		ReminderTime: reminderTimeString,
		DueDate:      ar.DueDate,
		TagID:        tagIdInt,
	}
}

type UpdateAssignmentRequest struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	DueDate      *int    `json:"dueDate,omitempty"`
	TagId        *string `json:"tagId,omitempty"`
	ReminderTime *[]int  `json:"reminderTime,omitempty" validate:"max=3,omitempty"`
}

func (ur *UpdateAssignmentRequest) UpdateAssignment(assignment *model.Assignment) {
	if ur.Name != nil {
		assignment.Name = *ur.Name
	}
	if ur.Description != nil {
		assignment.Description = *ur.Description
	}
	if ur.DueDate != nil {
		assignment.DueDate = *ur.DueDate
	}
	if ur.TagId != nil {
		tagIdInt, err := strconv.Atoi(*ur.TagId)
		if err == nil {
			assignment.TagID = tagIdInt
		}
	}
	if ur.ReminderTime != nil {
		assignment.ReminderTime = convertReminterTimeToString(*ur.ReminderTime)
	}
}

func convertReminterTimeToString(reminterTime []int) string {
	reminderTimeByte, _ := json.Marshal(reminterTime)
	return strings.Trim(string(reminderTimeByte), "[]")
}
