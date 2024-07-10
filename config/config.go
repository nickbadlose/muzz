package config

import (
	"fmt"
	"log"
	"os"
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
	cfgGeoIpEndpoint    = "GEOIP_ENDPOINT"
	cfgGeoIpAPIKey      = "GEOIP_API_KEY"
	cfgCachePassword    = "CACHE_PASSWORD"
	cfgCacheHost        = "CACHE_HOST"

	// defaults
	defaultEnv         = "development"
	defaultPort        = "3000"
	defaultJWTDuration = time.Hour * 6
)

var requiredEnv = []string{
	cfgDatabaseHost,
	cfgDatabaseUser,
	cfgDatabasePassword,
	cfgDatabase,
	cfgDomainName,
	cfgJWTDuration,
}

type Config struct{}

// Env retrieves the environment the app is running in.
func (cfg *Config) Env() string { return getEnv() }

// Port retrieves the port to run the server on.
func (cfg *Config) Port() string {
	port := viper.GetString(cfgPort)
	if port == "" {
		port = defaultPort
	}
	return fmt.Sprintf(":%s", port)
}

// DatabaseHost retrieves the host of the database to connect to.
func (cfg *Config) DatabaseHost() string { return viper.GetString(cfgDatabaseHost) }

// DatabaseUser retrieves the database user to authenticate with.
func (cfg *Config) DatabaseUser() string { return viper.GetString(cfgDatabaseUser) }

// DatabasePassword retrieves the database password to authenticate with.
func (cfg *Config) DatabasePassword() string { return viper.GetString(cfgDatabasePassword) }

// Database retrieves the database to connect to.
func (cfg *Config) Database() string { return viper.GetString(cfgDatabase) }

// LogLevel retrieves the application log level to run.
func (cfg *Config) LogLevel() string { return viper.GetString(cfgLogLevel) }

// DomainName retrieves the domain name that the server is hosted on.
func (cfg *Config) DomainName() string { return viper.GetString(cfgDomainName) }

// JWTDuration retrieves the expiry duration for JWTs.
func (cfg *Config) JWTDuration() time.Duration {
	dur := viper.GetDuration(cfgJWTDuration)
	if dur == 0 {
		dur = defaultJWTDuration
	}
	return dur
}

// JWTSecret retrieves the JWT secret to sign JWTs with.
func (cfg *Config) JWTSecret() string { return viper.GetString(cfgJWTSecret) }

// GeoIPAPIKey retrieves the API key for the geoip service..
func (cfg *Config) GeoIPAPIKey() string { return viper.GetString(cfgGeoIpAPIKey) }

// GeoIPEndpoint retrieves the endpoint for the geoip service..
func (cfg *Config) GeoIPEndpoint() string { return viper.GetString(cfgGeoIpEndpoint) }

// CacheHost retrieves the cache host to connect to.
func (cfg *Config) CacheHost() string { return viper.GetString(cfgCacheHost) }

// CachePassword retrieves the cache password to authenticate with.
func (cfg *Config) CachePassword() string { return viper.GetString(cfgCachePassword) }

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

// Load the environment into the viper package and returns a new Config interface to retrieve env vars. The env is
// configured from a root level "<environment>.env" file and then overwriting with environment variables.
func Load() (*Config, error) {
	env := getEnv()
	viper.AutomaticEnv()

	// env files aren't used in production.
	viper.AddConfigPath(".")
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
