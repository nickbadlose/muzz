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

// Config provides an interface to retrieve environment variables from.
type Config interface {
	// Env retrieves the environment the app is running in.
	Env() string
	// Port retrieves the port to run the server on.
	Port() string
	// DatabaseHost retrieves the host of the database to connect to.
	DatabaseHost() string
	// DatabaseUser retrieves the database user to authenticate with.
	DatabaseUser() string
	// DatabasePassword retrieves the database password to authenticate with.
	DatabasePassword() string
	// Database retrieves the database to connect to.
	Database() string
	// LogLevel retrieves the application log level to run.
	LogLevel() string
	// DomainName retrieves the domain name that the server is hosted on.
	DomainName() string
	// JWTDuration retrieves the expiry duration for JWTs.
	JWTDuration() time.Duration
	// JWTSecret retrieves the JWT secret to sign JWTs with.
	JWTSecret() string
}

type config struct{}

func (cfg *config) Env() string { return getEnv() }
func (cfg *config) Port() string {
	port := viper.GetString(cfgPort)
	if port == "" {
		port = defaultPort
	}
	return fmt.Sprintf(":%s", port)
}
func (cfg *config) DatabaseHost() string     { return viper.GetString(cfgDatabaseHost) }
func (cfg *config) DatabaseUser() string     { return viper.GetString(cfgDatabaseUser) }
func (cfg *config) DatabasePassword() string { return viper.GetString(cfgDatabasePassword) }
func (cfg *config) Database() string         { return viper.GetString(cfgDatabase) }
func (cfg *config) LogLevel() string         { return viper.GetString(cfgLogLevel) }
func (cfg *config) DomainName() string       { return viper.GetString(cfgDomainName) }
func (cfg *config) JWTDuration() time.Duration {
	dur := viper.GetDuration(cfgJWTDuration)
	if dur == 0 {
		dur = defaultJWTDuration
	}
	return dur
}
func (cfg *config) JWTSecret() string { return viper.GetString(cfgJWTSecret) }

// MustLoad calls Load and makes a call to log.Fatal if any required env vars haven't been set.
func MustLoad() Config {
	cfg := Load()

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
func Load() Config {
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
			log.Fatalf("reading env file: %s", err)
		}
	}

	return &config{}
}

func getEnv() string {
	env := os.Getenv(cfgEnvironment)
	if env == "" {
		env = defaultEnv
	}
	return env
}
