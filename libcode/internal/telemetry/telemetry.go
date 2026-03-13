// Package telemetry provides OpenTelemetry observability for libcode
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	globalTracer trace.Tracer
	initialized  bool
)

// Init initializes the OpenTelemetry tracing
func Init(serviceName, serviceVersion string, otlpEndpoint string) error {
	if initialized {
		return nil
	}

	// Create resource with service info
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter
	var exporter sdktrace.SpanExporter
	if otlpEndpoint != "" {
		exporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(otlpEndpoint),
		)
		if err != nil {
			return fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
	} else {
		// Use console exporter for development
		// exporter, _ = stdout.New(stdout.WithPrettyPrint())
		// For now, skip tracing if no endpoint
		initialized = true
		globalTracer = otel.Tracer(serviceName)
		return nil
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	globalTracer = otel.Tracer(serviceName)
	initialized = true

	return nil
}

// GetTracer returns the global tracer
func GetTracer() trace.Tracer {
	if !initialized {
		// Initialize with defaults
		_ = Init("libcode", "0.1.0-alpha", "")
	}
	return globalTracer
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name, trace.WithAttributes(attrs...))
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(attrs...))
}

// Shutdown flushes and closes the tracer provider
func Shutdown(ctx context.Context) error {
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}

// Helper functions for common spans

// WithSpan runs a function within a span
func WithSpan(ctx context.Context, name string, fn func(context.Context) error, attrs ...attribute.KeyValue) error {
	ctx, span := StartSpan(ctx, name, attrs...)
	defer span.End()

	if err := fn(ctx); err != nil {
		RecordError(ctx, err)
		return err
	}

	return nil
}

// Common attribute keys
const (
	// Provider attributes
	ProviderNameKey     = attribute.Key("provider.name")
	ProviderModelKey    = attribute.Key("provider.model")
	ProviderTokensKey   = attribute.Key("provider.tokens")

	// Tool attributes
	ToolNameKey         = attribute.Key("tool.name")
	ToolStatusKey       = attribute.Key("tool.status")
	ToolDurationKey     = attribute.Key("tool.duration_ms")

	// LSP attributes
	LSPServerKey        = attribute.Key("lsp.server")
	LSPMethodKey        = attribute.Key("lsp.method")
	LSPDurationKey      = attribute.Key("lsp.duration_ms")

	// MCP attributes
	MCPServerKey        = attribute.Key("mcp.server")
	MCPMethodKey        = attribute.Key("mcp.method")
	MCPDurationKey      = attribute.Key("mcp.duration_ms")

	// Session attributes
	SessionIDKey        = attribute.Key("session.id")
	SessionTitleKey      = attribute.Key("session.title")
	MessageCountKey     = attribute.Key("session.message_count")

	// Error attributes
	ErrorTypeKey        = attribute.Key("error.type")
	ErrorMessageKey     = attribute.Key("error.message")
)
