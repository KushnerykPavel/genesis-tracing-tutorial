package app

import (
	"context"
	"fmt"
	"github.com/KushnerykPavel/integration/internal/client"
	"github.com/KushnerykPavel/integration/internal/handler"
	"github.com/KushnerykPavel/integration/internal/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	c := &http.Client{Transport: http.DefaultTransport}

	pbClient := client.NewPrivatBankClient(c, appLogger, tracingProvider.Tracer("privatbank"))
	orderHandler := handler.NewOrderHandler(appLogger, tracingProvider.Tracer("order"), pbClient)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)

	r.Route("/api/v1/order", func(r chi.Router) {
		r.Get("/", orderHandler.ProcessOrder)
	})

	server := &http.Server{
		Addr:    app.cfg.Addr,
		Handler: otelhttp.NewHandler(r, "integration"),
	}

	errorg.Go(func() error {
		return server.ListenAndServe()
	})

	appLogger.Info(fmt.Sprintf("Server is running: %s", app.cfg.Addr))
	if err := errorg.Wait(); err != nil {
		return 1
	}

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Errorf("server shutdown error: %v", err)
		return 1
	}

	return 0
}
