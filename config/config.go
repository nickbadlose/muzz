package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

const (
	cfgEnvironment      = "ENVIRONMENT"
	cfgPort             = "PORT"
	cfgDatabaseHost     = "DATABASE_HOST"
	cfgDatabaseUser     = "DATABASE_USER"
	cfgDatabasePassword = "DATABASE_PASSWORD"
	cfgDatabase         = "DATABASE"
	cfgLogLevel         = "LOG_LEVEL"
	cfgDomainName       = "DOMAIN_NAME"
	cfgJWTDuration      = "JWT_DURATION"
	cfgJWTSecret        = "JWT_SECRET"
	cfgGeoIPEndpoint    = "GEOIP_ENDPOINT"
	// nolint:gosec // this is the env var names.
	cfgGeoIPAPIKey   = "GEOIP_API_KEY"
	cfgCacheHost     = "CACHE_HOST"
	cfgCachePassword = "CACHE_PASSWORD"
	cfgJaegerHost    = "JAEGER_HOST"
	cfgDebugEnabled  = "DEBUG_ENABLED"

	// default values.
	defaultEnv         = "development"
	defaultPort        = "3000"
	defaultJWTDuration = time.Hour * 6
)

// requiredEnv variables, MustLoad will make a call to os.Exit(1) if any of these are not set.
var requiredEnv = []string{
	cfgDatabaseHost,
	cfgDatabaseUser,
	cfgDatabasePassword,
	cfgDatabase,
	cfgDomainName,
	cfgJWTSecret,
	cfgGeoIPEndpoint,
	cfgGeoIPAPIKey,
	cfgCacheHost,
	cfgCachePassword,
	cfgJaegerHost,
}

// Config is the service which handles the retrieval of any environment variables and secrets.
type Config struct{}

// Env retrieves the environment the app is running in.
func (*Config) Env() string { return getEnv() }

// Port retrieves the port to run the server on.
func (*Config) Port() string {
	port := viper.GetString(cfgPort)
	if port == "" {
		port = defaultPort
	}
	return fmt.Sprintf(":%s", port)
}

// DatabaseHost retrieves the host of the database to connect to.
func (*Config) DatabaseHost() string { return mustString(cfgDatabaseHost) }

// DatabaseUser retrieves the database user to authenticate with.
func (*Config) DatabaseUser() string { return mustString(cfgDatabaseUser) }

// DatabasePassword retrieves the database password to authenticate with.
func (*Config) DatabasePassword() string { return mustString(cfgDatabasePassword) }

// Database retrieves the database to connect to.
func (*Config) Database() string { return mustString(cfgDatabase) }

// LogLevel retrieves the application log level to run.
func (*Config) LogLevel() string { return viper.GetString(cfgLogLevel) }

// DomainName retrieves the domain name that the server is hosted on.
func (*Config) DomainName() string { return mustString(cfgDomainName) }

// JWTDuration retrieves the expiry duration for JWTs.
func (*Config) JWTDuration() time.Duration {
	dur := viper.GetDuration(cfgJWTDuration)
	if dur == 0 {
		dur = defaultJWTDuration
	}
	return dur
}

// JWTSecret retrieves the JWT secret to sign JWTs with.
func (*Config) JWTSecret() string { return mustString(cfgJWTSecret) }

// GeoIPAPIKey retrieves the API key for the geoip service..
func (*Config) GeoIPAPIKey() string { return mustString(cfgGeoIPAPIKey) }

// GeoIPEndpoint retrieves the endpoint for the geoip service..
func (*Config) GeoIPEndpoint() string { return mustString(cfgGeoIPEndpoint) }

// CacheHost retrieves the cache host to connect to.
func (*Config) CacheHost() string { return mustString(cfgCacheHost) }

// CachePassword retrieves the cache password to authenticate with.
func (*Config) CachePassword() string { return mustString(cfgCachePassword) }

// JaegerHost retrieves the jaeger tracing host to deliver traces to.
func (*Config) JaegerHost() string { return mustString(cfgJaegerHost) }

// DebugEnabled retrieves the application debug configuration.
func (*Config) DebugEnabled() bool { return viper.GetBool(cfgDebugEnabled) }

// MustLoad calls Load and makes a call to log.Fatal if any required env vars haven't been set.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("loading config: %s", err)
	}

	// check if required env vars are set.
	for _, key := range requiredEnv {
		value := viper.Get(key)
		if value == "" {
			log.Fatalf("environment variable '%s' not set", key)
		}
	}

	return cfg
}

// Load the environment into the viper package and returns a new Config service to retrieve env vars. The env is
// configured from a root level "<environment>.env" file and then overwriting with set environment variables.
func Load() (*Config, error) {

	// TODO pass in paths ...string param, which adds config paths to search for env file.
	//  This way we can keep env files in related packages

	env := getEnv()
	viper.AutomaticEnv()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("unable to get caller information")
	}

	// env files aren't used in production.
	viper.AddConfigPath(fmt.Sprintf("%s/../", path.Dir(filename)))
	viper.SetConfigName(env)
	viper.SetConfigType("env")
	err := viper.ReadInConfig()
	if err != nil {
		// this is the viper recommended way to check if the error is from no env file found.
		// https://github.com/spf13/viper?tab=readme-ov-file#reading-config-files
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	return &Config{}, nil
}

func getEnv() string {
	env := os.Getenv(cfgEnvironment)
	if env == "" {
		env = defaultEnv
	}
	return env
}

func mustString(k string) string {
	v := viper.GetString(k)
	if v == "" {
		log.Fatalf("could not find config value for: %v", k)
	}

	return v
}
