package config

import (
	"errors"
	"fmt"
	"os"
)

var errEnvVarNotFound error = errors.New("environment variable not found")

const (
	apiPortEnvKey   = "API_PORT"
	ethNodeEnvKey   = "ETH_NODE_URL"
	dbConnEnvKey    = "DB_CONNECTION_URL"
	jwtSecretEnvKey = "JWT_SECRET"
)

type Config struct {
	Port               string
	NodeURL            string
	DBConnectionString string
	JWTSecret          string
}

func NewConfig() (*Config, error) {
	port, ok := os.LookupEnv(apiPortEnvKey)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errEnvVarNotFound, apiPortEnvKey)
	}

	nodeURL, ok := os.LookupEnv(ethNodeEnvKey)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errEnvVarNotFound, ethNodeEnvKey)
	}

	dbConn, ok := os.LookupEnv(dbConnEnvKey)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errEnvVarNotFound, dbConnEnvKey)
	}

	jwtSecret, ok := os.LookupEnv(jwtSecretEnvKey)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errEnvVarNotFound, jwtSecretEnvKey)
	}

	return &Config{
		Port:               port,
		NodeURL:            nodeURL,
		DBConnectionString: dbConn,
		JWTSecret:          jwtSecret,
	}, nil
}
