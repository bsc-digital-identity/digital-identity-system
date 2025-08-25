package appbuilder

import (
	"pkg-common/logger"
	"pkg-common/rabbitmq"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Application struct {
	Logger         *logger.Logger
	Addr           string
	Conn           *amqp.Connection
	WorkerServices []rabbitmq.WorkerService
	Engine         *gin.Engine
}

type ApplicationInterface interface {
	Start()
}

func (a *Application) Start() {
	a.Logger.Info("Starting Application runtime...")

	for _, ws := range a.WorkerServices {
		a.Logger.Infof("Starting %s WorkerService", ws.GetServiceName())
		go ws.StartService()
	}

	a.Logger.Infof("REST API is now listening on: %s", a.Addr)
	a.Engine.Run(a.Addr)
}
