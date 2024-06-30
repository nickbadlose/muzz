package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	defaultHost         = "3000"
	cfgPort             = "port"
	cfgDatabaseHost     = "database_host"
	cfgDatabaseUser     = "database_user"
	cfgDatabasePassword = "database_password"
	cfgDatabase         = "database"
	cfgLogLevel         = "log_level"
)

type Config struct {
	// Port to run the server on.
	Port string
	// DatabaseHost is host to connect to.
	DatabaseHost string
	// DatabaseUser to authenticate with.
	DatabaseUser string
	// DatabasePassword to authenticate with.
	DatabasePassword string
	// Database to connect to.
	Database string
	// LogLevel represents the log level of the application.
	LogLevel string
}

// MustLoad returns a new *Config struct, configured form the environment.
// It panics if any configurations are not provided.
func MustLoad() *Config {
	port := getString(cfgPort)
	if port == "" {
		port = defaultHost
	}

	return &Config{
		Port:             fmt.Sprintf(":%s", port),
		DatabaseHost:     mustString(cfgDatabaseHost),
		DatabaseUser:     mustString(cfgDatabaseUser),
		DatabasePassword: mustString(cfgDatabasePassword),
		Database:         mustString(cfgDatabase),
		LogLevel:         getString(cfgLogLevel),
	}
}

func getString(key string) string {
	return os.Getenv(strings.ToUpper(key))
}

func mustString(key string) string {
	value := os.Getenv(strings.ToUpper(key))
	if value == "" {
		log.Fatalf("environment variable %s not set", strings.ToUpper(key))
	}

	return value
}
