package tracing

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	serviceName    = "unknown-service"
	shutdownTracer = func(context.Context) error { return nil }
)

type Config interface {
	CollectorEndpoint() string
	ServiceName() string
	Environment() string
	ServiceVersion() string
}

func InitTracer(ctx context.Context, cfg Config) error {
	serviceName = cfg.ServiceName()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("deployment.environment", cfg.Environment()),
			attribute.String("service.version", cfg.ServiceVersion()),
		),
	)
	if err != nil {
		return err
	}

	providerOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	}

	if endpoint := strings.TrimSpace(cfg.CollectorEndpoint()); endpoint != "" {
		exporter, err := otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return err
		}
		providerOptions = append(providerOptions, sdktrace.WithBatcher(exporter))
	}

	provider := sdktrace.NewTracerProvider(providerOptions...)
	otel.SetTracerProvider(provider)
	shutdownTracer = provider.Shutdown

	return nil
}

func ShutdownTracer(ctx context.Context) error {
	return shutdownTracer(ctx)
}

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(serviceName).Start(ctx, name, opts...)
}

func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().TraceID().String()
}
