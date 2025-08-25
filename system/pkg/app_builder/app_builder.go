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

type appBuilder[T utilities.JsonConfigObj[U], U AppConfig] struct {
	logger         *logger.Logger
	config         U
	conn           *amqp.Connection
	workerServices []rabbitmq.WorkerService
	routes         []rest.Route
	engine         *gin.Engine
}

type AppBuilderInterface[T utilities.JsonConfigObj[U], U AppConfig] interface {
	InitLogger(loggerArgs logger.GlobalLoggerConfig) AppBuilderInterface[T, U]
	LoadConfig(configPath string) AppBuilderInterface[T, U]
	InitRabbitmqConnection() AppBuilderInterface[T, U]
	InitRabbitmqRegistries() AppBuilderInterface[T, U]
	AddWorkerServices(workerServices ...rabbitmq.WorkerService) AppBuilderInterface[T, U]
	AddSwagger() AppBuilderInterface[T, U]
	AddGinRoutes(routes ...rest.Route) AppBuilderInterface[T, U]
	InitGinRouter() AppBuilderInterface[T, U]
	Build() ApplicationInterface
}

func New[T utilities.JsonConfigObj[U], U AppConfig]() AppBuilderInterface[T, U] {
	return &appBuilder[T, U]{}
}

func (a *appBuilder[T, U]) InitLogger(loggerArgs logger.GlobalLoggerConfig) AppBuilderInterface[T, U] {
	logger.InitDefaultLogger(loggerArgs)
	a.logger = logger.Default()
	a.logger.Info("Logger initialized")

	return a
}

func (a *appBuilder[T, U]) LoadConfig(filePath string) AppBuilderInterface[T, U] {
	a.logger.Infof("Preparing to load config from %s ...", filePath)
	jsonConfig, err := utilities.ReadConfig[T, U](filePath)
	if err != nil {
		a.logger.Error(err, "Failed to load config")
		panic(err)
	}

	a.config = jsonConfig
	a.logger.Info("Config successfully loaded.")
	return a
}

func (a *appBuilder[T, U]) InitRabbitmqConnection() AppBuilderInterface[T, U] {
	a.logger.Info("Preparing to connect to Rabbitmq server...")
	rabbitmqConfig := a.config.GetRabbitmqConfig()
	conn, err := rabbitmq.ConnectToRabbitmq(
		rabbitmqConfig.User,
		rabbitmqConfig.Password,
	)
	if err != nil {
		panic(err)
	}

	a.conn = conn
	a.logger.Info("Connection with Rabbitmq server established")

	return a
}

func (a *appBuilder[T, U]) InitRabbitmqRegistries() AppBuilderInterface[T, U] {
	a.logger.Info("Initializing Rabbitmq registries from config")
	rabbitmqConf := a.config.GetRabbitmqConfig()

	rabbitmq.InitializeConsumerRegistry(a.conn, rabbitmqConf.ConsumersConfig)
	rabbitmq.InitializePublisherRegistry(a.conn, rabbitmqConf.PublishersConfig)
	a.logger.Info("Successfully initialized Rabbitmq registries from config")

	return a
}

func (a *appBuilder[T, U]) AddWorkerServices(workerServices ...rabbitmq.WorkerService) AppBuilderInterface[T, U] {
	a.logger.Info("Adding Worker Services to Application...")
	a.workerServices = append(a.workerServices, workerServices...)
	return a
}

func (a *appBuilder[T, U]) AddGinRoutes(routes ...rest.Route) AppBuilderInterface[T, U] {
	a.logger.Info("Adding Gin REST API routes to Application...")
	a.routes = append(a.routes, routes...)
	return a
}

func (a *appBuilder[T, U]) AddSwagger() AppBuilderInterface[T, U] {
	a.logger.Info("Adding SwaggerUI...")
	a.routes = append(a.routes, rest.NewRoute(
		rest.GET,
		"swagger",
		"*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	))

	return a
}

func (a *appBuilder[T, U]) InitGinRouter() AppBuilderInterface[T, U] {
	a.logger.Info("Initializing Gin Router...")
	router := gin.Default()

	groups := map[string]*gin.RouterGroup{}
	a.logger.Info("Registering REST API routes...")
	for _, r := range a.routes {
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
			group.PUT(r.Path, r.HandlerFunc)
		case rest.PATCH:
			group.PATCH(r.Path, r.HandlerFunc)
		default:
			a.logger.Warnf("Unrecognized HTTP method: %s", r.Method)
		}
	}

	a.engine = router
	a.logger.Info("Successfully registered REST API routes.")
	return a
}

func (a *appBuilder[T, U]) Build() ApplicationInterface {
	return &application{
		Logger:         a.logger,
		Addr:           fmt.Sprintf("0.0.0.0:%d", a.config.GetRestApiPort()),
		Conn:           a.conn,
		WorkerServices: a.workerServices,
		Engine:         a.engine,
	}
}
