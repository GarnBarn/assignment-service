package main

import (
	"fmt"

	"github.com/GarnBarn/common-go/database"
	"github.com/GarnBarn/common-go/httpserver"
	"github.com/GarnBarn/common-go/logger"
	"github.com/GarnBarn/common-go/proto"
	"github.com/GarnBarn/gb-assignment-service/config"
	"github.com/GarnBarn/gb-assignment-service/handler"
	"github.com/GarnBarn/gb-assignment-service/repository"
	"github.com/GarnBarn/gb-assignment-service/service"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Dial gRPC Server
	grpcConn, err := grpc.Dial(appConfig.TAG_GRPC_SERVER, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("Can't connect gRPC Server: ", err)
	}
	defer grpcConn.Close()

	tagClient := proto.NewTagClient(grpcConn)

	// Create the required dependentices
	validate := validator.New()

	// Create repositroy
	assignmentRepository := repository.NewAssignmentRepository(db)

	// Create service
	assignmentService := service.NewAssignmentService(tagClient, assignmentRepository)

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
