package app

import (
	"context"
	"fmt"
	"github.com/KushnerykPavel/consumer/internal/consumer"
	"github.com/KushnerykPavel/consumer/internal/integration"
	"github.com/KushnerykPavel/consumer/internal/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

type App struct {
	cfg *Config
}

func New(cfg *Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (app *App) Run(appCtx context.Context) int {
	errorg := errgroup.Group{}
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	logger, _ := config.Build()
	defer logger.Sync()
	appLogger := logger.Sugar().With("service", app.cfg.ServiceName)
	ctx, cancelFunc := context.WithCancel(appCtx)
	defer cancelFunc()

	tracingProvider, shutdown, err := tracing.SetupOTelSDK(ctx, app.cfg.ServiceName, app.cfg.OtelURL)
	if err != nil {
		appLogger.Errorf("failed to setup tracing: %v", err)
		return 1
	}
	defer shutdown(ctx)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{app.cfg.KafkaURL},
		Topic:   app.cfg.TopicName,
		GroupID: "consumer",
	})

	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	c := integration.NewClient(client, appLogger, tracingProvider.Tracer("integration-client"), app.cfg.IntegrationURL)

	cons := consumer.NewOrderConsumer(reader, appLogger, tracingProvider.Tracer("order-consumer"), c)

	errorg.Go(func() error {
		return cons.ConsumeMessage(ctx)
	})

	r := chi.NewRouter()
	r.Use(middleware.RealIP)

	server := &http.Server{
		Addr:    app.cfg.Addr,
		Handler: r,
	}

	errorg.Go(func() error {
		return server.ListenAndServe()
	})

	appLogger.Info(fmt.Sprintf("Server is running: %s", app.cfg.Addr))
	if err := errorg.Wait(); err != nil {
		return 1
	}

	if err := reader.Close(); err != nil {
		appLogger.Errorf("kafka.Close: %v", err)
		return 1

	}

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Errorf("server shutdown error: %v", err)
		return 1
	}

	return 0
}
