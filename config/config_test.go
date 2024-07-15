package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig_success(t *testing.T) {
	err := os.Setenv("ENVIRONMENT", "test_config")
	require.NoError(t, err)

	cfg, err := Load()
	require.NoError(t, err)

	require.Equal(t, "test_config", cfg.Env())
	require.Equal(t, ":3000", cfg.Port())
	require.Equal(t, "test_database_user", cfg.DatabaseUser())
	require.Equal(t, "test_database_password", cfg.DatabasePassword())
	require.Equal(t, "test_database", cfg.Database())
	require.Equal(t, "test_database_host", cfg.DatabaseHost())
	require.Equal(t, "test_log_level", cfg.LogLevel())
	require.Equal(t, "test_domain_name", cfg.DomainName())
	require.Equal(t, time.Duration(43200000000000), cfg.JWTDuration())
	require.Equal(t, "test_jwt_secret", cfg.JWTSecret())
	require.Equal(t, "test_geoip_api_key", cfg.GeoIPAPIKey())
	require.Equal(t, "test_geoip_endpoint", cfg.GeoIPEndpoint())
	require.Equal(t, "test_cache_host", cfg.CacheHost())
	require.Equal(t, "test_cache_password", cfg.CachePassword())
	require.Equal(t, "test_jaeger_host", cfg.JaegerHost())
	require.Equal(t, true, cfg.DebugEnabled())
}
