package rest

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestSetupTelemetry(t *testing.T) {
	registry := prometheus.NewRegistry()

	t.Run("successfully creates telemetry", func(t *testing.T) {
		telemetry, err := SetupTelemetry(registry)
		require.NoError(t, err)
		require.NotNil(t, telemetry)
		require.NotNil(t, telemetry.MeterProvider())
	})

	t.Run("shutdowns cleanly", func(t *testing.T) {
		telemetry, err := SetupTelemetry(registry)
		require.NoError(t, err)

		err = telemetry.Shutdown(context.Background())
		require.NoError(t, err)
	})
}
