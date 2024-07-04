package queue

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type HeadersCarrier map[string]interface{}

func (a HeadersCarrier) Get(key string) string {
	v, ok := a[key]
	if !ok {
		return ""
	}
	return v.(string)
}

func (a HeadersCarrier) Set(key string, value string) {
	a[key] = value
}

func (a HeadersCarrier) Keys() []string {
	i := 0
	r := make([]string, len(a))

	for k := range a {
		r[i] = k
		i++
	}

	return r
}

func InjectPHeaders(ctx context.Context) []kafka.Header {
	h := make(HeadersCarrier)
	otel.GetTextMapPropagator().Inject(ctx, h)
	headers := make([]kafka.Header, 0, len(h))
	for k, v := range h {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v.(string))})
	}
	return headers
}

func ExtractPHeaders(ctx context.Context, headers []kafka.Header) context.Context {
	h := make(HeadersCarrier)
	for _, header := range headers {
		h[header.Key] = string(header.Value)
	}
	return otel.GetTextMapPropagator().Extract(ctx, h)
}
