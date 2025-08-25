package appbuilder

import (
	"fmt"
	"pkg-common/logger"
	"pkg-common/rabbitmq"
	"pkg-common/rest"
	"pkg-common/utilities"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type AppConfig interface {
	GetLoggerConfig() logger.LoggerConfig
	GetRabbitmqConfig() rabbitmq.RabbitmqConfig
	GetRestApiPort() uint16
}

type AppBuilder[T utilities.JsonConfigObj[U], U AppConfig] struct {
	Logger         *logger.Logger
	Config         U
	Conn           *amqp.Connection
	WorkerServices []rabbitmq.WorkerService
	Routes         []rest.Route
	Engine         *gin.Engine
}

type AppBuilderInterface[T utilities.JsonConfigObj[U], U AppConfig] interface {
	InitLogger(loggerArgs []struct{ Key, Value string }) *AppBuilder[T, U]
	LoadConfig(configPath string) *AppBuilder[T, U]
	InitRabbitmqConnection() *AppBuilder[T, U]
	InitRabbitmqRegistries() *AppBuilder[T, U]
	AddWorkerServices(workerServices ...rabbitmq.WorkerService)
	AddSwagger()
	AddGinRoute(routes ...rest.Route) *AppBuilder[T, U]
	InitGinRouter() *AppBuilder[T, U]
	Build() *Application
}

func New[T utilities.JsonConfigObj[U], U AppConfig]() *AppBuilder[T, U] {
	return &AppBuilder[T, U]{}
}

func (a *AppBuilder[T, U]) InitLogger(loggerArgs logger.GlobalLoggerConfig) *AppBuilder[T, U] {
	logger.InitDefaultLogger(loggerArgs)
	a.Logger = logger.Default()
	a.Logger.Info("Logger initialized")

	return a
}

func (a *AppBuilder[T, U]) LoadConfig(filePath string) *AppBuilder[T, U] {
	a.Logger.Infof("Preparing to load config from %s ...", filePath)
	jsonConfig, err := utilities.ReadConfig[T, U](filePath)
	if err != nil {
		a.Logger.Error(err, "Failed to load config")
		panic(err)
	}

	a.Config = jsonConfig
	a.Logger.Info("Config sucessfully loaded.")
	return a
}

func (a *AppBuilder[T, U]) InitRabbitmqConnection() *AppBuilder[T, U] {
	a.Logger.Info("Preparing to connect to Rabbitmq server...")
	rabbitmqConfig := a.Config.GetRabbitmqConfig()
	conn, err := rabbitmq.ConnectToRabbitmq(
		rabbitmqConfig.User,
		rabbitmqConfig.Password,
	)
	if err != nil {
		panic(err)
	}

	a.Conn = conn
	a.Logger.Info("Connection with Rabbitmq server established")

	return a
}

func (a *AppBuilder[T, U]) InitRabbitmqRegistries() *AppBuilder[T, U] {
	a.Logger.Info("Initializing Rabbitmq registries from config")
	rabbitmqConf := a.Config.GetRabbitmqConfig()

	rabbitmq.InitializeConsumerRegistry(a.Conn, rabbitmqConf.ConsumersConfig)
	rabbitmq.InitializePublisherRegistry(a.Conn, rabbitmqConf.PublishersConfig)
	a.Logger.Info("Sucessfully initialized Rabbitmq registries from config")

	return a
}

func (a *AppBuilder[T, U]) AddWorkerServices(workerServices ...rabbitmq.WorkerService) *AppBuilder[T, U] {
	a.Logger.Info("Adding Worker Services to Application...")
	a.WorkerServices = append(a.WorkerServices, workerServices...)
	return a
}

func (a *AppBuilder[T, U]) AddGinRoutes(routes ...rest.Route) *AppBuilder[T, U] {
	a.Logger.Info("Adding Gin REST API routes to Application...")
	a.Routes = append(a.Routes, routes...)
	return a
}

func (a *AppBuilder[T, U]) AddSwagger() *AppBuilder[T, U] {
	a.Logger.Info("Adding SwaggerUI...")
	a.Routes = append(a.Routes, rest.NewRoute(
		rest.GET,
		"swagger",
		"*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	))

	return a
}

func (a *AppBuilder[T, U]) InitGinRouter() *AppBuilder[T, U] {
	a.Logger.Info("Initializng Gin Router...")
	router := gin.Default()

	groups := map[string]*gin.RouterGroup{}
	a.Logger.Info("Registering REST API routes...")
	for _, r := range a.Routes {
		if _, exists := groups[r.Group]; !exists {
			groups[r.Group] = router.Group("/" + r.Group)
		}

		group := groups[r.Group]

		switch r.Method {
		case rest.GET:
			group.GET(r.Path, r.HandlerFunc)
		case rest.POST:
			group.POST(r.Path, r.HandlerFunc)
		case rest.PUT:
			group.GET(r.Path, r.HandlerFunc)
		case rest.PATCH:
			group.GET(r.Path, r.HandlerFunc)
		default:
			a.Logger.Warnf("Unrecoginzed HTTP method: %s", r.Method)
		}
	}

	a.Engine = router
	a.Logger.Info("Sucessfully registered REST API routes.")
	return a
}

func (a *AppBuilder[T, U]) Build() *Application {
	return &Application{
		Logger:         a.Logger,
		Addr:           fmt.Sprintf("0.0.0.0:%d", a.Config.GetRestApiPort()),
		Conn:           a.Conn,
		WorkerServices: a.WorkerServices,
		Engine:         a.Engine,
	}
}
