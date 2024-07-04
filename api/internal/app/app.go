package app

import (
	"context"
	"fmt"
	"github.com/KushnerykPavel/api/internal/handler"
	"github.com/KushnerykPavel/api/internal/queue"
	"github.com/KushnerykPavel/api/internal/repo"
	"github.com/KushnerykPavel/api/internal/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
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

	db, err := sqlx.Connect("postgres", app.cfg.DBURL)
	if err != nil {
		appLogger.Errorf("sqlx.Connect: %v", err)
		return 1
	}

	conn, err := kafka.DialLeader(ctx, "tcp", app.cfg.KafkaURL, app.cfg.TopicName, 0)
	if err != nil {
		appLogger.Errorf("kafka.Connect: %v", err)
		return 1
	}

	ordersQueue := queue.NewOrdersQueue(conn, appLogger, tracingProvider.Tracer("orders-queue"))

	repository := repo.New(appLogger, db, tracingProvider.Tracer("repository"))

	urlHandler := handler.NewOrderHandler(
		appLogger,
		tracingProvider.Tracer("order-handler"),
		repository,
		ordersQueue,
	)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)

	r.Route("/api/v1/order", func(r chi.Router) {
		r.Post("/", urlHandler.CreateOrder)
	})

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

	if err := conn.Close(); err != nil {
		appLogger.Errorf("kafka.Close: %v", err)
		return 1

	}

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Errorf("server shutdown error: %v", err)
		return 1
	}
	return 0
}
