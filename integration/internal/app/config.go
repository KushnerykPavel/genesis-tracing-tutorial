package app

type Config struct {
	Addr        string `envconfig:"ADDR" required:"true"`
	ServiceName string `envconfig:"SERVICE_NAME" required:"true"`
	OtelURL     string `envconfig:"OTEL_URL" required:"true"`
}
