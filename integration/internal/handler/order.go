package handler

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

type externalClient interface {
	GetEURRate(ctx context.Context) (float64, error)
}

type OrderHandler struct {
	l              *zap.SugaredLogger
	tracer         trace.Tracer
	externalClient externalClient
}

func NewOrderHandler(l *zap.SugaredLogger, tracerProvider trace.Tracer, e externalClient) *OrderHandler {
	return &OrderHandler{
		l:              l,
		tracer:         tracerProvider,
		externalClient: e,
	}
}

func (o *OrderHandler) ProcessOrder(w http.ResponseWriter, r *http.Request) {
	spanCtx, span := o.tracer.Start(r.Context(), "ProcessOrder")
	defer span.End()

	_, err := o.externalClient.GetEURRate(spanCtx)
	if err != nil {
		o.l.Errorw("failed to get eur rate", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
