package tracer

import (
	"context"
	"fmt"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/nickbadlose/muzz/internal/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type Config interface {
	Env() string
	JaegerHost() string
}

// New configures the global OpenTelemetry tracer and returns an error if it fails.
func New(ctx context.Context, cfg Config, serviceName string) (*tracesdk.TracerProvider, error) {
	logger.Debug(
		ctx,
		"setting up Jaeger tracer",
		zap.String("host", cfg.JaegerHost()),
		zap.String("env", cfg.Env()),
	)

	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(fmt.Sprintf("http://%s/api/traces", cfg.JaegerHost()))),
	)
	if err != nil {
		return nil, err
	}

	// tracerProvider returns an OpenTelemetry TracerProvider configured to use
	// the Jaeger exporter that will send spans to the provided exporter. The returned
	// TracerProvider will also use a Resource configured with all the information
	// about the application.
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			attribute.String("environment", cfg.Env()),
			attribute.String("ID", uuid.NewString()),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}
