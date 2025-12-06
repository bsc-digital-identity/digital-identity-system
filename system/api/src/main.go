package main

import (
	"api/src/database"
	"api/src/docs"
	"api/src/identity"
	logaudit "api/src/log_audit"
	"api/src/middleware"
	"api/src/outbox"
	zkpfailed "api/src/zkp/failed"
	zkpresult "api/src/zkp/results"
	"api/src/zkprequest"
	"fmt"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
	"pkg-common/rabbitmq"
	"pkg-common/rest"
	"pkg-common/utilities"
	"time"
)

// @title           Digital Identity System API
// @version         1.0
// @description     API to manage identities and verify ZKP proofs
// @host localhost:9000
// @BasePath /v1
func main() {

	var zkpHandler *zkprequest.Handler

	lanHost := utilities.ResolveLanHost()
	apiBaseURL := fmt.Sprintf("http://%s:9000", lanHost)
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:9000", lanHost)

	var logAuditHandler *logaudit.LogAuditHandler

	appbuilder.New[ApiConfigJson, ApiConfig]().
		InitLogger(logger.GlobalLoggerConfig{}).
		ResolveEnvironment().
		LoadConfig("config.json").
		WithOption(func(a *appbuilder.AppBuilder[ApiConfigJson, ApiConfig]) {
			// ----- DATABASE + MIGRATIONS -----
			database.ConnectToDatabase(a)
			database.RunMigrations(true)

			// ----- ZKP SERVICE FIXED -----
			svc := zkprequest.NewService(
				&zkprequest.InMemoryStore{},
				func(s *zkprequest.Service) {
					s.Audience = apiBaseURL
				},
				func(s *zkprequest.Service) {
					s.ResponseURI = apiBaseURL + "/v1/presentations/verify"
				},
				func(s *zkprequest.Service) {
					s.TTL = 5 * time.Minute
				},
			)

			zkpHandler = zkprequest.NewHandler(svc)

			// ----- LOG AUDIT SERVICE -----
			logAuditRepo := logaudit.NewLogAuditRepository()
			logAuditService := logaudit.NewLogAuditService(logAuditRepo)
			logAuditHandler = logaudit.NewLogAuditHandler(logAuditService)
		}).

		// ----- RABBITMQ -----
		InitRabbitmqConnection().
		InitRabbitmqRegistries().
		WithOption(func(a *appbuilder.AppBuilder[ApiConfigJson, ApiConfig]) {
			// ----- RABBITMQ LOGGING SINK -----
			logPublisher := rabbitmq.GetPublisher("LogPublisher")
			loggerInstance := logger.Default()
			logSink := rabbitmq.CreateRabbitmqLoggerSink(logPublisher)
			logger.AddSinkToLoggerInstance(loggerInstance, logSink)
		}).
		// ----- WORKERS -----
		AddWorkerServices(
			zkpfailed.NewZeroKnowledgeProofFailedHandler(),
			zkpresult.NewZeroKnowledgeProofHandler(),
			outbox.NewOutboxWorker(),
			logaudit.NewLogSinkWorker(),
		).

		// ----- CORS (ONE GOOD MIDDLEWARE) -----
		AddGinMiddleware(
			rest.NewMiddleware("*", middleware.CORSMiddleware()),
			rest.NewMiddleware("v1/internal", rest.InternalAuthMiddleware()),
		).

		// ----- ROUTES -----
		AddGinRoutes(
			rest.NewRoute(rest.POST, "v1", "identity", identity.NewHandler().CreateIdentity),
			rest.NewRoute(rest.GET, "v1", "identity/:id", identity.NewHandler().GetIdentity),
			rest.NewRoute(rest.POST, "v1", "identity/verify", identity.NewHandler().QueueVerification),

			// ZKP Presentation Request
			rest.NewRoute(rest.POST, "v1", "presentations/create", zkpHandler.CreatePresentation),

			rest.NewRoute(rest.POST, "v1", "presentations/verify", zkpHandler.VerifyPresentation),
			rest.NewRoute(rest.GET, "v1", "presentations/:request_id", zkpHandler.ShowPresentation),
			rest.NewRoute(rest.GET, "v1", "presentations/:request_id/descriptor", zkpHandler.Descriptor),
			rest.NewRoute(rest.GET, "v1", "presentations/:request_id/status", zkpHandler.Status),
			rest.NewRoute(rest.GET, "v1", "presentations/:request_id/result", zkpHandler.Result),

			// NEW: schema JSON pod hashem
			rest.NewRoute(rest.GET, "v1", "schemas/:hash", zkpHandler.Schema),
			rest.NewRoute(rest.GET, "v1", "artifacts/:hash/vk", zkpHandler.GetVK),
			rest.NewRoute(rest.GET, "v1", "artifacts/:hash/pk", zkpHandler.GetPK),

			// LOG AUDIT ROUTES:
			rest.NewRoute(rest.GET, "v1", "logs", logAuditHandler.GetLogEntries),
			rest.NewRoute(rest.GET, "v1", "logs/service/:service", logAuditHandler.GetLogEntriesByService),
			rest.NewRoute(rest.GET, "v1", "logs/level/:level", logAuditHandler.GetLogEntriesByLevel),

			// DEV ONLY:
			rest.NewRoute(rest.POST, "v1", "presentations/verify-blocking", zkpHandler.VerifyBlocking),
		).
		AddSwagger().
		InitGinRouter().
		Build().
		Start()
}
