package app

type Config struct {
	Addr        string `envconfig:"ADDR" required:"true"`
	DBURL       string `envconfig:"DB_URL" required:"true"`
	ServiceName string `envconfig:"SERVICE_NAME" required:"true"`
	OtelURL     string `envconfig:"OTEL_URL" required:"true"`

	KafkaURL  string `envconfig:"KAFKA_URL" required:"true"`
	TopicName string `envconfig:"TOPIC_NAME" required:"true"`
}
