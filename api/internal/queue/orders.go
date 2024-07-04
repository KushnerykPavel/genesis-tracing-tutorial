package queue

import (
	"context"
	"encoding/json"
	"github.com/KushnerykPavel/api/internal/entity"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"time"
)

type OrdersQueue struct {
	conn   *kafka.Conn
	tracer trace.Tracer
	l      *zap.SugaredLogger
}

type order struct {
	CustomerID string  `json:"customer_id"`
	OrderID    int     `json:"order_id"`
	Price      float64 `json:"price"`
	CreatedAt  string  `json:"created_at"`
}

func NewOrdersQueue(conn *kafka.Conn, l *zap.SugaredLogger, t trace.Tracer) *OrdersQueue {
	return &OrdersQueue{
		conn:   conn,
		tracer: t,
		l:      l,
	}
}

func (q *OrdersQueue) PublishOrder(ctx context.Context, o entity.Order) error {
	spanCtx, span := q.tracer.Start(ctx, "create-order-queue")
	defer span.End()

	ord := order{
		CustomerID: o.CustomerID,
		Price:      o.Price,
		CreatedAt:  o.CreatedAt.Format(time.DateTime),
		OrderID:    o.OrderID,
	}
	body, err := json.Marshal(ord)
	if err != nil {
		span.RecordError(err)
		q.l.Errorw("failed to marshal order", "error", err)
		return err
	}
	headers := InjectPHeaders(spanCtx)

	message := kafka.Message{
		Value:   body,
		Headers: headers,
	}
	q.l.With("order_id", ord.OrderID).Info("publishing order")
	_, err = q.conn.WriteMessages([]kafka.Message{
		message,
	}...)
	return err
}
