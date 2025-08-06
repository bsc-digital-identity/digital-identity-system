package rabbitmq

import "pkg-common/utilities"

type RabbimqConfigJson struct {
	User             string                         `json:"user"`
	Password         string                         `json:"password"`
	PublishersConfig []RabbitmqPublishersConfigJson `json:"publishers"`
	ConsumersConfig  []RabbitmqConsumerConfigJson   `json:"consumers"`
}

type RabbitmqConfig struct {
	User             string
	Password         string
	PublishersConfig []RabbitmqPublishersConfig
	ConsumersConfig  []RabbitmqConsumerConfig
}

func (rcj RabbimqConfigJson) ConvertToDomain() RabbitmqConfig {
	return RabbitmqConfig{
		User:     rcj.User,
		Password: rcj.Password,
		PublishersConfig: utilities.ConvertJsonArrayToDomain[
			RabbitmqPublishersConfigJson,
			RabbitmqPublishersConfig,
		](rcj.PublishersConfig),
		ConsumersConfig: utilities.ConvertJsonArrayToDomain[
			RabbitmqConsumerConfigJson,
			RabbitmqConsumerConfig,
		](rcj.ConsumersConfig),
	}
}

type RabbitmqPublishersConfigJson struct {
	PublisherAlias string `json:"publisher_alias"`
	Exchange       string `json:"exchange"`
	RoutingKey     string `json:"routing_key"`
}

type RabbitmqPublishersConfig struct {
	PublisherAlias PublisherAlias
	Exchange       string
	RoutingKey     string
}

func (rpcj RabbitmqPublishersConfigJson) ConvertToDomain() RabbitmqPublishersConfig {
	return RabbitmqPublishersConfig{
		PublisherAlias: PublisherAlias(rpcj.PublisherAlias),
		Exchange:       rpcj.Exchange,
		RoutingKey:     rpcj.RoutingKey,
	}
}

type RabbitmqConsumerConfigJson struct {
	ConsumerAlias string `json:"consumer_alias"`
	ConsumerTag   string `json:"consumer_tag"`
	QueueName     string `json:"queue_name"`
}

type RabbitmqConsumerConfig struct {
	ConsumerAlias ConsumerAlias
	ConsumerTag   string
	QueueName     string
}

func (rccj RabbitmqConsumerConfigJson) ConvertToDomain() RabbitmqConsumerConfig {
	return RabbitmqConsumerConfig{
		ConsumerAlias: ConsumerAlias(rccj.ConsumerAlias),
		QueueName:     rccj.QueueName,
		ConsumerTag:   rccj.ConsumerTag,
	}
}
