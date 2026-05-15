package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const TraceIDHeader = "x-trace-id"

func UnaryServerInterceptor(service string) grpc.UnaryServerInterceptor {
	tracer := otel.GetTracerProvider().Tracer(service)
	propagator := otel.GetTextMapPropagator()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		ctx = propagator.Extract(ctx, metadataCarrier(md))
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		ctx = AddTraceIDToResponse(ctx)
		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
		}

		return resp, err
	}
}

func UnaryClientInterceptor(service string) grpc.UnaryClientInterceptor {
	tracer := otel.GetTracerProvider().Tracer(service)
	propagator := otel.GetTextMapPropagator()

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		outgoing := extractOutgoingMetadata(ctx)
		carrier := metadataCarrier(outgoing)
		propagator.Inject(ctx, carrier)
		ctx = metadata.NewOutgoingContext(ctx, metadata.MD(carrier))

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}

func GetTraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().TraceID().String()
}

func AddTraceIDToResponse(ctx context.Context) context.Context {
	traceID := GetTraceIDFromContext(ctx)
	if traceID == "" {
		return ctx
	}

	md := extractOutgoingMetadata(ctx)
	md.Set(TraceIDHeader, traceID)
	return metadata.NewOutgoingContext(ctx, md)
}

func extractOutgoingMetadata(ctx context.Context) metadata.MD {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return metadata.New(nil)
	}
	return md.Copy()
}
