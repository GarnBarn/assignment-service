package main

import (
	"fmt"

	"github.com/GarnBarn/assignment-service/config"
	"github.com/GarnBarn/common-go/database"
	"github.com/GarnBarn/common-go/httpserver"
	"github.com/GarnBarn/common-go/logger"
	"github.com/GarnBarn/gb-assignment-service/handler"
	"github.com/GarnBarn/gb-assignment-service/repository"
	"github.com/GarnBarn/gb-assignment-service/service"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var (
	appConfig config.Config
)

func init() {
	appConfig = config.Load()
	logger.InitLogger(logger.Config{
		Env: appConfig.Env,
	})

}

func main() {
	// Create the http server
	httpServer := httpserver.NewHttpServer()

	db, err := database.Conn(appConfig.MYSQL_CONNECTION_STRING)
	if err != nil {
		logrus.Fatalln("Can't connect to database: ", err)
		return
	}

	// Create the required dependentices
	validate := validator.New()

	// Create repositroy
	assignmentRepository := repository.NewAssignmentRepository(db)

	// Create service
	assignmentService := service.NewAssignmentService(assignmentRepository)

	// Create Handler
	assignmentHandler := handler.NewAssignmentHandler(*validate, assignmentService)

	// Router
	router := httpServer.Group("/api/v1")

	// Assignment
	assignmentRouter := router.Group("/assignment")
	assignmentRouter.POST("/", assignmentHandler.CreateAssignment)
	assignmentRouter.GET("/", assignmentHandler.GetAllAssignment)

	logrus.Info("Listening and serving HTTP on :", appConfig.HTTP_SERVER_PORT)
	httpServer.Run(fmt.Sprint(":", appConfig.HTTP_SERVER_PORT))
}
