package main

import (
	"api/src/database"
	_ "api/src/docs"
	"api/src/identity"
	"api/src/middleware"
	"api/src/outbox"
	zkpfailed "api/src/zkp/failed"
	zkpresult "api/src/zkp/results"
	"api/src/zkprequest"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
	"pkg-common/rest"
	"time"
)

// @title           Digital Identity System API
// @version         1.0
// @description     API to manage identities and verify ZKP proofs
// @host localhost:9000
// @BasePath /v1
func main() {

	var zkpHandler *zkprequest.Handler

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
					// Correct URL - NO double http://
					s.Audience = "http://192.168.8.107:9000"
				},
				func(s *zkprequest.Service) {
					// Correct URL - NO double http://
					s.ResponseURI = "http://192.168.8.107:9000/v1/presentations/verify"
				},
				func(s *zkprequest.Service) {
					s.TTL = 5 * time.Minute
				},
			)

			zkpHandler = zkprequest.NewHandler(svc)
		}).

		// ----- RABBITMQ -----
		InitRabbitmqConnection().
		InitRabbitmqRegistries().

		// ----- WORKERS -----
		AddWorkerServices(
			zkpfailed.NewZeroKnowledgeProofFailedHandler(),
			zkpresult.NewZeroKnowledgeProofHandler(),
			outbox.NewOutboxWorker(),
		).

		// ----- CORS (ONE GOOD MIDDLEWARE) -----
		AddGinMiddleware(
			rest.NewMiddleware("v1", middleware.CORSMiddleware()),
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

			// DEV ONLY:
			rest.NewRoute(rest.POST, "v1", "presentations/mock-verify", zkpHandler.MockVerify),

			rest.NewRoute(rest.POST, "v1", "presentations/verify-blocking", zkpHandler.VerifyBlocking),
		).
		AddSwagger().
		InitGinRouter().
		Build().
		Start()
}
