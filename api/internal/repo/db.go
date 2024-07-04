package repo

import (
	"context"
	"github.com/KushnerykPavel/api/internal/entity"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"time"
)

type Repository struct {
	l      *zap.SugaredLogger
	db     *sqlx.DB
	tracer trace.Tracer
}

func New(l *zap.SugaredLogger, db *sqlx.DB, t trace.Tracer) *Repository {
	return &Repository{l: l.With("module", "repo"), db: db, tracer: t}
}

func (r *Repository) CreateOrder(ctx context.Context, order entity.Order) (int, error) {
	spanCtx, span := r.tracer.Start(ctx, "create-order-repo")
	defer span.End()
	span.SetAttributes(
		attribute.String("customer_id", order.CustomerID),
		attribute.String("order_created_at", order.CreatedAt.Format(time.DateTime)),
	)
	var id int
	err := r.db.QueryRowContext(spanCtx,
		"INSERT INTO orders (customer_id, price, created_at) VALUES ($1, $2, $3) RETURNING id",
		order.CustomerID, order.Price, order.CreatedAt.Format(time.DateTime)).Scan(&id)
	if err != nil {
		span.RecordError(err)
		r.l.Errorw("failed to create order", "error", err)
		return 0, err
	}
	return id, nil
}
