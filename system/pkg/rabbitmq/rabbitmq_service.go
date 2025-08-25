package rabbitmq

type WorkerService interface {
	StartService()
	GetServiceName() string
}
