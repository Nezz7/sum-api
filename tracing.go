package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func InitTracer() (func(context.Context) error, error) {
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		return nil, fmt.Errorf("JAEGER_ENDPOINT var not found")
	}

	// https://opentelemetry.io/docs/languages/go/exporters/
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(jaegerEndpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithURLPath("/v1/traces"),
	)
	if err != nil {
		return nil, err
	}

	// https://opentelemetry.io/docs/languages/go/resources/
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("sum-api"),
			semconv.ServiceVersion("1.0.0"),
			semconv.ServerPort(8080),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		// https://opentelemetry.io/docs/languages/go/sampling/
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	slog.Info("Init tracing using Jaeger: %s" + jaegerEndpoint)

	return tp.Shutdown, nil
}
