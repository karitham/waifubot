package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/rest/api"
)

func TestOgenTelemetryIntegration(t *testing.T) {
	registry := prometheus.NewRegistry()

	telemetry, err := SetupTelemetry(registry)
	require.NoError(t, err)
	defer telemetry.Shutdown(context.Background())

	srv, err := api.NewServer(
		api.UnimplementedHandler{},
		api.WithMeterProvider(telemetry.MeterProvider()),
	)
	require.NoError(t, err)

	t.Run("requests are tracked by ogen telemetry", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/123", nil)
		rw := httptest.NewRecorder()

		srv.ServeHTTP(rw, req)

		metrics, err := registry.Gather()
		require.NoError(t, err)
		require.NotEmpty(t, metrics)
	})

	t.Run("metrics include request counts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/123", nil)
		rw := httptest.NewRecorder()

		srv.ServeHTTP(rw, req)

		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, m := range metrics {
			t.Logf("Metric: %s", m.GetName())
			for _, metric := range m.GetMetric() {
				t.Logf("  Labels: %+v", metric.GetLabel())
			}
		}
	})
}
