package handler

import (
	"context"
	"github.com/KushnerykPavel/api/internal/entity"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type repository interface {
	CreateOrder(context.Context, entity.Order) (int, error)
}

type queue interface {
	PublishOrder(context.Context, entity.Order) error
}

type OrderHandler struct {
	l      *zap.SugaredLogger
	tracer trace.Tracer
	repo   repository
	q      queue
}

func NewOrderHandler(l *zap.SugaredLogger, tracerProvider trace.Tracer, r repository, q queue) *OrderHandler {
	return &OrderHandler{
		l:      l,
		tracer: tracerProvider,
		repo:   r,
		q:      q,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	spanCtx, span := h.tracer.Start(context.Background(), "create-order-handler")
	defer span.End()

	orderReq := &orderRequest{}
	if err := render.Bind(r, orderReq); err != nil {
		span.SetStatus(codes.Error, "parse url request error")
		span.RecordError(err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	orderEntity := entity.Order{
		CustomerID: orderReq.CustomerID,
		Price:      orderReq.Price,
		CreatedAt:  time.Now(),
	}
	orderID, err := h.repo.CreateOrder(spanCtx, orderEntity)
	if err != nil {
		span.SetStatus(codes.Error, "create order error")
		span.RecordError(err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	orderEntity.OrderID = orderID

	if err = h.q.PublishOrder(spanCtx, orderEntity); err != nil {
		span.SetStatus(codes.Error, "publish order error")
		span.RecordError(err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}
