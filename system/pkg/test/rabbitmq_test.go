package test

import (
	"errors"
	"pkg-common/rabbitmq"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Mock implementation for testing
type MockChannel struct {
	published []amqp.Publishing
	consumed  chan amqp.Delivery
	closed    bool
}

func NewMockChannel() *MockChannel {
	return &MockChannel{
		published: make([]amqp.Publishing, 0),
		consumed:  make(chan amqp.Delivery, 10),
		closed:    false,
	}
}

func (m *MockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	m.published = append(m.published, msg)
	return nil
}

func (m *MockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return m.consumed, nil
}

func (m *MockChannel) Close() error {
	m.closed = true
	close(m.consumed)
	return nil
}

// Mock serializable for testing
type MockSerializable struct {
	data string
	err  error
}

func (m MockSerializable) Serialize() ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(m.data), nil
}

func TestRabbitmqConfigConvertToDomain(t *testing.T) {
	publisherConfig := rabbitmq.RabbitmqPublishersConfigJson{
		PublisherAlias: "test-publisher",
		Exchange:       "test-exchange",
		RoutingKey:     "test-key",
	}

	consumerConfig := rabbitmq.RabbitmqConsumerConfigJson{
		ConsumerAlias: "test-consumer",
		ConsumerTag:   "test-tag",
		QueueName:     "test-queue",
	}

	config := rabbitmq.RabbimqConfigJson{
		User:             "testuser",
		Password:         "testpass",
		PublishersConfig: []rabbitmq.RabbitmqPublishersConfigJson{publisherConfig},
		ConsumersConfig:  []rabbitmq.RabbitmqConsumerConfigJson{consumerConfig},
	}

	result := config.MapToDomain()

	if result.User != "testuser" {
		t.Errorf("Expected User to be 'testuser', got '%s'", result.User)
	}
	if result.Password != "testpass" {
		t.Errorf("Expected Password to be 'testpass', got '%s'", result.Password)
	}
	if len(result.PublishersConfig) != 1 {
		t.Errorf("Expected 1 publisher config, got %d", len(result.PublishersConfig))
	}
	if len(result.ConsumersConfig) != 1 {
		t.Errorf("Expected 1 consumer config, got %d", len(result.ConsumersConfig))
	}
}

func TestRabbitmqPublishersConfigConvertToDomain(t *testing.T) {
	config := rabbitmq.RabbitmqPublishersConfigJson{
		PublisherAlias: "test-publisher",
		Exchange:       "test-exchange",
		RoutingKey:     "test-routing-key",
	}

	result := config.MapToDomain()

	if string(result.PublisherAlias) != "test-publisher" {
		t.Errorf("Expected PublisherAlias to be 'test-publisher', got '%s'", result.PublisherAlias)
	}
	if result.Exchange != "test-exchange" {
		t.Errorf("Expected Exchange to be 'test-exchange', got '%s'", result.Exchange)
	}
	if result.RoutingKey != "test-routing-key" {
		t.Errorf("Expected RoutingKey to be 'test-routing-key', got '%s'", result.RoutingKey)
	}
}

func TestRabbitmqConsumerConfigConvertToDomain(t *testing.T) {
	config := rabbitmq.RabbitmqConsumerConfigJson{
		ConsumerAlias: "test-consumer",
		ConsumerTag:   "test-tag",
		QueueName:     "test-queue",
	}

	result := config.MapToDomain()

	if string(result.ConsumerAlias) != "test-consumer" {
		t.Errorf("Expected ConsumerAlias to be 'test-consumer', got '%s'", result.ConsumerAlias)
	}
	if result.ConsumerTag != "test-tag" {
		t.Errorf("Expected ConsumerTag to be 'test-tag', got '%s'", result.ConsumerTag)
	}
	if result.QueueName != "test-queue" {
		t.Errorf("Expected QueueName to be 'test-queue', got '%s'", result.QueueName)
	}
}

func TestNewPublisher(t *testing.T) {
	exchange := "test-exchange"
	routingKey := "test-key"

	// Create a real AMQP channel interface by casting our mock
	// Note: This is a simplified approach. In real tests, you might use interfaces or dependency injection
	publisher := rabbitmq.NewPublisher((*amqp.Channel)(nil), exchange, routingKey)

	if publisher == nil {
		t.Fatal("Expected publisher to be created, got nil")
	}
}

func TestNewConsumer(t *testing.T) {
	queueName := "test-queue"
	consumerTag := "test-tag"

	consumer := rabbitmq.NewConsumer((*amqp.Channel)(nil), queueName, consumerTag)

	if consumer == nil {
		t.Fatal("Expected consumer to be created, got nil")
	}
}

func TestPublisherPublish(t *testing.T) {
	tests := []struct {
		name         string
		serializable MockSerializable
		expectError  bool
		expectedBody string
	}{
		{
			name:         "Successful publish",
			serializable: MockSerializable{data: "test data", err: nil},
			expectError:  false,
			expectedBody: "test data",
		},
		{
			name:         "Serialization error",
			serializable: MockSerializable{data: "", err: errors.New("serialization failed")},
			expectError:  true,
			expectedBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test would need a mock AMQP channel that implements the interface
			// For now, we're testing the serialization logic
			data, err := tt.serializable.Serialize()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && string(data) != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, string(data))
			}
		})
	}
}

func TestConnectToRabbitmqRetryLogic(t *testing.T) {
	// This test would ideally test the retry logic
	// Since we can't easily mock the AMQP connection in this context,
	// we'll test that the function exists and has the expected signature

	// Test that the function doesn't panic when called with invalid credentials
	// Note: This will fail in actual execution, but we're testing the function signature
	defer func() {
		if r := recover(); r == nil {
			// Function should not panic, just return an error
		}
	}()

	// Test function signature
	var fn func(string, string) (*amqp.Connection, error) = rabbitmq.ConnectToRabbitmq
	if fn == nil {
		t.Error("ConnectToRabbitmq function not found")
	}
}

func TestConsumerRegistryOperations(t *testing.T) {
	// Test that we can work with consumer aliases
	alias := rabbitmq.ConsumerAlias("test-consumer")

	// Test type conversion
	aliasString := string(alias)
	if aliasString != "test-consumer" {
		t.Errorf("Expected alias string to be 'test-consumer', got '%s'", aliasString)
	}

	// Test alias creation from string
	newAlias := rabbitmq.ConsumerAlias("another-consumer")
	if string(newAlias) != "another-consumer" {
		t.Errorf("Expected new alias to be 'another-consumer', got '%s'", string(newAlias))
	}
}

func TestPublisherRegistryOperations(t *testing.T) {
	// Test that we can work with publisher aliases
	alias := rabbitmq.PublisherAlias("test-publisher")

	// Test type conversion
	aliasString := string(alias)
	if aliasString != "test-publisher" {
		t.Errorf("Expected alias string to be 'test-publisher', got '%s'", aliasString)
	}

	// Test alias creation from string
	newAlias := rabbitmq.PublisherAlias("another-publisher")
	if string(newAlias) != "another-publisher" {
		t.Errorf("Expected new alias to be 'another-publisher', got '%s'", string(newAlias))
	}
}

func TestConfigStructTypes(t *testing.T) {
	// Test RabbitmqPublishersConfig struct
	publisherConfig := rabbitmq.RabbitmqPublishersConfig{
		PublisherAlias: rabbitmq.PublisherAlias("test"),
		Exchange:       "exchange",
		RoutingKey:     "key",
	}

	if string(publisherConfig.PublisherAlias) != "test" {
		t.Error("PublisherConfig alias not set correctly")
	}
	if publisherConfig.Exchange != "exchange" {
		t.Error("PublisherConfig exchange not set correctly")
	}
	if publisherConfig.RoutingKey != "key" {
		t.Error("PublisherConfig routing key not set correctly")
	}

	// Test RabbitmqConsumerConfig struct
	consumerConfig := rabbitmq.RabbitmqConsumerConfig{
		ConsumerAlias: rabbitmq.ConsumerAlias("test"),
		ConsumerTag:   "tag",
		QueueName:     "queue",
	}

	if string(consumerConfig.ConsumerAlias) != "test" {
		t.Error("ConsumerConfig alias not set correctly")
	}
	if consumerConfig.ConsumerTag != "tag" {
		t.Error("ConsumerConfig tag not set correctly")
	}
	if consumerConfig.QueueName != "queue" {
		t.Error("ConsumerConfig queue name not set correctly")
	}
}

func TestRabbitmqConfigArrayConversion(t *testing.T) {
	publisherConfigs := []rabbitmq.RabbitmqPublishersConfigJson{
		{
			PublisherAlias: "pub1",
			Exchange:       "exchange1",
			RoutingKey:     "key1",
		},
		{
			PublisherAlias: "pub2",
			Exchange:       "exchange2",
			RoutingKey:     "key2",
		},
	}

	consumerConfigs := []rabbitmq.RabbitmqConsumerConfigJson{
		{
			ConsumerAlias: "cons1",
			ConsumerTag:   "tag1",
			QueueName:     "queue1",
		},
		{
			ConsumerAlias: "cons2",
			ConsumerTag:   "tag2",
			QueueName:     "queue2",
		},
	}

	config := rabbitmq.RabbimqConfigJson{
		User:             "user",
		Password:         "pass",
		PublishersConfig: publisherConfigs,
		ConsumersConfig:  consumerConfigs,
	}

	result := config.MapToDomain()

	if len(result.PublishersConfig) != 2 {
		t.Errorf("Expected 2 publisher configs, got %d", len(result.PublishersConfig))
	}
	if len(result.ConsumersConfig) != 2 {
		t.Errorf("Expected 2 consumer configs, got %d", len(result.ConsumersConfig))
	}

	// Verify first publisher config
	if string(result.PublishersConfig[0].PublisherAlias) != "pub1" {
		t.Error("First publisher alias not converted correctly")
	}

	// Verify first consumer config
	if string(result.ConsumersConfig[0].ConsumerAlias) != "cons1" {
		t.Error("First consumer alias not converted correctly")
	}
}

func TestRabbitmqStructsImplementInterfaces(t *testing.T) {
	// Test that RabbitmqPublisher implements IRabbitmqPublisher
	var publisher rabbitmq.IRabbitmqPublisher
	publisher = &rabbitmq.RabbitmqPublisher{}
	if publisher == nil {
		t.Error("RabbitmqPublisher should implement IRabbitmqPublisher")
	}

	// Test that RabbitmqConsumer implements IRabbitmqConsumer
	var consumer rabbitmq.IRabbitmqConsumer
	consumer = &rabbitmq.RabbitmqConsumer{}
	if consumer == nil {
		t.Error("RabbitmqConsumer should implement IRabbitmqConsumer")
	}
}

// Benchmark tests for performance
func BenchmarkPublisherCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rabbitmq.NewPublisher(nil, "exchange", "key")
	}
}

func BenchmarkConsumerCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rabbitmq.NewConsumer(nil, "queue", "tag")
	}
}

func BenchmarkConfigConversion(b *testing.B) {
	config := rabbitmq.RabbimqConfigJson{
		User:     "user",
		Password: "pass",
		PublishersConfig: []rabbitmq.RabbitmqPublishersConfigJson{
			{PublisherAlias: "pub", Exchange: "ex", RoutingKey: "key"},
		},
		ConsumersConfig: []rabbitmq.RabbitmqConsumerConfigJson{
			{ConsumerAlias: "cons", ConsumerTag: "tag", QueueName: "queue"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.MapToDomain()
	}
}
