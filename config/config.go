package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	// defaultHost is the default host/port for the web app.
	defaultHost = "3000"
	// cfgHost is the configuration key for the host/port to bind to.
	cfgHost = "host"
)

type Config struct {
	Host string
}

// MustLoad returns a new *Config struct, configured form the environment.
// It panics if any configurations are not provided.
func MustLoad() *Config {
	host := getString(cfgHost)
	if host == "" {
		host = defaultHost
	}

	return &Config{Host: fmt.Sprintf(":%s", host)}
}

func getString(key string) string {
	return os.Getenv(strings.ToUpper(key))
}
