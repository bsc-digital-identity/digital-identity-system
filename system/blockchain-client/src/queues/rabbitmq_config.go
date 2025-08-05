package queues

import "pkg-common/utilities"

type RabbimqConfigJson struct {
	Exchanges []RabbitmqExchangeConfigJson `json:"exchanges"`
	Queues    []RabbitmqQueueConfigJson    `json:"queues"`
}

type RabbitmqConfig struct {
	Exchanges []RabbitmqExchangeConfig
	Queues    []RabbitmqQueueConfig
}

func (rcj RabbimqConfigJson) ConvertToDomain() RabbitmqConfig {
	return RabbitmqConfig{
		Exchanges: utilities.ConvertJsonArrayToDomain[RabbitmqExchangeConfigJson, RabbitmqExchangeConfig](rcj.Exchanges),
		Queues:    utilities.ConvertJsonArrayToDomain[RabbitmqQueueConfigJson, RabbitmqQueueConfig](rcj.Queues),
	}
}

type RabbitmqExchangeConfigJson struct {
	ExchangeName string `json:"exchange_name"`
	ExchangeType string `json:"exchange_type"`
}

type RabbitmqExchangeConfig struct {
	ExchangeName string
	ExchangeType RabbitmqExchangeType
}

func (recj RabbitmqExchangeConfigJson) ConvertToDomain() RabbitmqExchangeConfig {
	return RabbitmqExchangeConfig{
		ExchangeName: recj.ExchangeName,
		ExchangeType: RabbitmqExchangeType(recj.ExchangeType),
	}
}

type RabbitmqQueueConfigJson struct {
	QueueName       string `json:"queue_name"`
	RoutingKey      string `json:"routing_key"`
	ExchangeBinding string `json:"exchange_binding"`
	Durable         bool   `json:"durable"`
	Exclusive       bool   `json:"exclusive"`
}

type RabbitmqQueueConfig struct {
	QueueName       string
	RoutingKey      string
	ExchangeBinding string
	Durable         bool
	Exclusive       bool
}

func (rqcj RabbitmqQueueConfigJson) ConvertToDomain() RabbitmqQueueConfig {
	return RabbitmqQueueConfig{
		QueueName:       rqcj.QueueName,
		RoutingKey:      rqcj.RoutingKey,
		ExchangeBinding: rqcj.ExchangeBinding,
		Durable:         rqcj.Durable,
		Exclusive:       rqcj.Exclusive,
	}
}
