package integration

import (
	"context"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

type Client struct {
	c              *http.Client
	tracer         trace.Tracer
	l              *zap.SugaredLogger
	integrationURL string
}

func NewClient(c *http.Client, l *zap.SugaredLogger, t trace.Tracer, intURL string) *Client {
	return &Client{
		c:              c,
		tracer:         t,
		l:              l,
		integrationURL: intURL,
	}
}

func (cl *Client) CallIntegration(ctx context.Context) error {
	spanCtx, span := cl.tracer.Start(ctx, "call-integration")
	defer span.End()

	url := cl.integrationURL + "/api/v1/order"
	req, err := http.NewRequestWithContext(spanCtx, http.MethodGet, url, nil)
	if err != nil {
		span.SetStatus(codes.Error, "failed to read body")
		span.RecordError(err)
		cl.l.Errorw("failed to create request", "error", err)
		return err
	}

	_, err = cl.c.Do(req)

	return err
}
