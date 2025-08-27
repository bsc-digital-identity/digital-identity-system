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
	conn           *amqp.Connection
	workerServices []rabbitmq.WorkerService
	routes         []rest.Route
	engine         *gin.Engine
}

type AppBuilderInterface[T utilities.JsonConfigObj[U], U AppConfig] interface {
	InitLogger(loggerArgs logger.GlobalLoggerConfig) AppBuilderInterface[T, U]
	ResolveEnvironment() AppBuilderInterface[T, U]
	LoadConfig(configPath string) AppBuilderInterface[T, U]
	WithOption(func(*AppBuilder[T, U])) AppBuilderInterface[T, U]
	InitRabbitmqConnection() AppBuilderInterface[T, U]
	InitRabbitmqRegistries() AppBuilderInterface[T, U]
	AddWorkerServices(workerServices ...rabbitmq.WorkerService) AppBuilderInterface[T, U]
	AddSwagger() AppBuilderInterface[T, U]
	AddGinRoutes(routes ...rest.Route) AppBuilderInterface[T, U]
	InitGinRouter() AppBuilderInterface[T, U]
	Build() ApplicationInterface
}

func New[T utilities.JsonConfigObj[U], U AppConfig]() AppBuilderInterface[T, U] {
	return &AppBuilder[T, U]{}
}

func (a *AppBuilder[T, U]) InitLogger(loggerArgs logger.GlobalLoggerConfig) AppBuilderInterface[T, U] {
	logger.InitDefaultLogger(loggerArgs)
	a.Logger = logger.Default()
	a.Logger.Info("Logger initialized")

	return a
}

func (a *AppBuilder[T, U]) LoadConfig(filePath string) AppBuilderInterface[T, U] {
	a.Logger.Infof("Preparing to load config from %s ...", filePath)
	jsonConfig, err := utilities.ReadConfig[T, U](filePath)
	if err != nil {
		a.Logger.Error(err, "Failed to load config")
		panic(err)
	}

	a.Config = jsonConfig
	a.Logger.Info("Config successfully loaded.")
	return a
}

func (a *AppBuilder[T, U]) ResolveEnvironment() AppBuilderInterface[T, U] {
	// TODO: implement later
	return a
}

func (a *AppBuilder[T, U]) InitRabbitmqConnection() AppBuilderInterface[T, U] {
	a.Logger.Info("Preparing to connect to Rabbitmq server...")
	rabbitmqConfig := a.Config.GetRabbitmqConfig()
	conn, err := rabbitmq.ConnectToRabbitmq(
		rabbitmqConfig.User,
		rabbitmqConfig.Password,
	)
	if err != nil {
		panic(err)
	}

	a.conn = conn
	a.Logger.Info("Connection with Rabbitmq server established")

	return a
}

func (a *AppBuilder[T, U]) InitRabbitmqRegistries() AppBuilderInterface[T, U] {
	a.Logger.Info("Initializing Rabbitmq registries from config")
	rabbitmqConf := a.Config.GetRabbitmqConfig()

	rabbitmq.InitializeConsumerRegistry(a.conn, rabbitmqConf.ConsumersConfig)
	rabbitmq.InitializePublisherRegistry(a.conn, rabbitmqConf.PublishersConfig)
	a.Logger.Info("Successfully initialized Rabbitmq registries from config")

	return a
}

func (a *AppBuilder[T, U]) AddWorkerServices(workerServices ...rabbitmq.WorkerService) AppBuilderInterface[T, U] {
	a.Logger.Info("Adding Worker Services to Application...")
	a.workerServices = append(a.workerServices, workerServices...)
	return a
}

func (a *AppBuilder[T, U]) AddGinRoutes(routes ...rest.Route) AppBuilderInterface[T, U] {
	a.Logger.Info("Adding Gin REST API routes to Application...")
	a.routes = append(a.routes, routes...)
	return a
}

func (a *AppBuilder[T, U]) AddSwagger() AppBuilderInterface[T, U] {
	a.Logger.Info("Adding SwaggerUI...")
	a.routes = append(a.routes, rest.NewRoute(
		rest.GET,
		"swagger",
		"*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	))

	return a
}

func (a *AppBuilder[T, U]) InitGinRouter() AppBuilderInterface[T, U] {
	a.Logger.Info("Initializing Gin Router...")
	router := gin.Default()

	groups := map[string]*gin.RouterGroup{}
	a.Logger.Info("Registering REST API routes...")
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
			a.Logger.Warnf("Unrecognized HTTP method: %s", r.Method)
		}
	}

	a.engine = router
	a.Logger.Info("Successfully registered REST API routes.")
	return a
}

func (a *AppBuilder[T, U]) Build() ApplicationInterface {
	return &application{
		Logger:         a.Logger,
		Addr:           fmt.Sprintf("0.0.0.0:%d", a.Config.GetRestApiPort()),
		Conn:           a.conn,
		WorkerServices: a.workerServices,
		Engine:         a.engine,
	}
}

func (a *AppBuilder[T, U]) WithOption(fn func(*AppBuilder[T, U])) AppBuilderInterface[T, U] {
	fn(a)
	return a
}
