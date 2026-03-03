package rest

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Telemetry struct {
	meterProvider metric.MeterProvider
	shutdown      func(context.Context) error
}

const appName = "waifubot_api"

func SetupTelemetry(registerer prometheus.Registerer) (*Telemetry, error) {
	appResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(appName),
	)

	exp, err := otelprometheus.New(
		otelprometheus.WithRegisterer(registerer),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(appResource),
		sdkmetric.WithReader(exp),
	)

	return &Telemetry{
		meterProvider: mp,
		shutdown: func(ctx context.Context) error {
			_ = mp.ForceFlush(ctx)
			return mp.Shutdown(ctx)
		},
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down telemetry")
	if err := t.shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown telemetry", "error", err)
		return err
	}
	slog.Info("Telemetry shutdown complete")
	return nil
}

func (t *Telemetry) MeterProvider() metric.MeterProvider {
	return t.meterProvider
}
