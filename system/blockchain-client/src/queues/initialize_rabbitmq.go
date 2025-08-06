package queues

type RabbitmqExchangeType string

func (ret RabbitmqExchangeType) String() string {
	return string(ret)
}

// ConnectToRabbitmq connects with retries
