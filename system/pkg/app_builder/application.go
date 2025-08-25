package appbuilder

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"pkg-common/logger"
	"pkg-common/rabbitmq"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type application struct {
	Logger         *logger.Logger
	Addr           string
	Conn           *amqp.Connection
	WorkerServices []rabbitmq.WorkerService
	Engine         *gin.Engine
}

type ApplicationInterface interface {
	Start()
}

func (a *application) Start() {
	a.Logger.Info("Starting Application runtime...")

	for _, ws := range a.WorkerServices {
		a.Logger.Infof("Starting %s WorkerService", ws.GetServiceName())
		go ws.StartService()
	}

	a.Logger.Info("Starting server...")
	srv := &http.Server{
		Addr:    a.Addr,
		Handler: a.Engine,
	}

	go func() {
		a.Logger.Infof("REST API is now listening on: %s", a.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatalf(err, "Listen errors: ")
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
	<-exit

	a.Logger.Info("Recevied shutdown signal, closing application...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.Logger.Errorf(err, "Server forced to shutdown.")
	}

	if a.Conn != nil {
		a.Conn.Close()
		a.Logger.Info("Connection to Rabbitmq server closed.")
	}

	a.Logger.Info("Application closed gracefully.")
}
