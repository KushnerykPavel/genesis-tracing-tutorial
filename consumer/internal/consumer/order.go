package consumer

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type order struct {
	CustomerID string  `json:"customer_id"`
	OrderID    int     `json:"order_id"`
	Price      float64 `json:"price"`
	CreatedAt  string  `json:"created_at"`
}

type integrationService interface {
	CallIntegration(ctx context.Context) error
}

type OrderConsumer struct {
	reader             *kafka.Reader
	tracer             trace.Tracer
	l                  *zap.SugaredLogger
	integrationService integrationService
}

func NewOrderConsumer(reader *kafka.Reader, l *zap.SugaredLogger, t trace.Tracer, in integrationService) *OrderConsumer {
	return &OrderConsumer{
		reader:             reader,
		tracer:             t,
		l:                  l.With("service", "order-consumer"),
		integrationService: in,
	}
}

func (c *OrderConsumer) ConsumeMessage(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.l.Errorw("failed to read message", "error", err)
			return err
		}
		messageCtx := ExtractPHeaders(ctx, m.Headers)

		c.l.Info("received message", "message", string(m.Value))
		if err := c.processMessage(messageCtx, m.Value); err != nil {
			c.l.Errorw("failed to process message", "error", err)
		}
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.l.Errorw("failed to commit message", "error", err)
		}
	}
}

func (c *OrderConsumer) processMessage(ctx context.Context, message []byte) error {
	spanCtx, span := c.tracer.Start(ctx, "consume-order-message")
	defer span.End()

	var o order
	if err := json.Unmarshal(message, &o); err != nil {
		span.RecordError(err)
		c.l.Errorw("failed to unmarshal order", "error", err)
		return err
	}

	span.SetAttributes(
		attribute.Int("order_id", o.OrderID),
		attribute.String("customer_id", o.CustomerID),
	)

	return c.integrationService.CallIntegration(spanCtx)
}
