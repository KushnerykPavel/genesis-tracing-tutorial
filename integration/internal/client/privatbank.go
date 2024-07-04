package client

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const url = "https://api.privatbank.ua/p24api/pubinfo?json=null&exchange=null&coursid=5"

type PrivatBank struct {
	c      *http.Client
	l      *zap.SugaredLogger
	tracer trace.Tracer
}

func NewPrivatBankClient(c *http.Client, l *zap.SugaredLogger, t trace.Tracer) *PrivatBank {
	return &PrivatBank{
		c:      c,
		l:      l,
		tracer: t,
	}
}

func (p *PrivatBank) GetEURRate(ctx context.Context) (float64, error) {
	spanCtx, span := p.tracer.Start(ctx, "GetEURRate")
	req, err := http.NewRequestWithContext(spanCtx, http.MethodGet, url, http.NoBody)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create request")
		span.RecordError(err)
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := p.c.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, "failed to do request")
		span.RecordError(err)
		return 0, fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		span.SetStatus(codes.Error, "failed to read body")
		span.RecordError(err)
		return 0, fmt.Errorf("failed to read body: %w", err)
	}
	p.l.Info("sucessfully receive body: ", string(body))
	return 0, nil
}
