package config

import (
	"os"
	"strconv"
	"time"
)

const (
	MinimumHeight = 33
	MinimumWidth  = 120
)

var (
	debug  bool = false
	logCli bool = false

	serverHost        string = "localhost"
	serverPortSSH     string = "2222"
	serverPortWeb     string = "8080"
	webAdminKey       string
	webAllowedOrigins string = ""

	database    string = "farkle"
	databaseURI string = "mongodb://localhost:27017"

	playerTimeoutMinutes int = 30 // Default 30 minute timeout

	Version string = "dev"
)

func Setup() {
	debugStr := os.Getenv("ASCII_ARCADE_DEBUG")
	if debugStr != "" {
		debug = debugStr == "true" || debugStr == "1"
	}

	logCliStr := os.Getenv("ASCII_ARCADE_LOG_CLI")
	if logCliStr != "" {
		logCli = logCliStr == "true" || logCliStr == "1"
	}

	hostStr := os.Getenv("ASCII_ARCADE_HOST")
	if hostStr != "" {
		serverHost = hostStr
	}
	portStr := os.Getenv("ASCII_ARCADE_SSH_PORT")
	if portStr != "" {
		serverPortSSH = portStr
	}
	webPortStr := os.Getenv("ASCII_ARCADE_WEB_PORT")
	if webPortStr != "" {
		serverPortWeb = webPortStr
	}
	webAdminKey = os.Getenv("ASCII_ARCADE_WEB_ADMIN_KEY")
	if webAdminKey == "" {
		webAdminKey = "supersecretkey"
	}
	webAllowedOrigins = os.Getenv("ASCII_ARCADE_WEB_ALLOWED_ORIGINS")

	dbStr := os.Getenv("ASCII_ARCADE_DB")
	if dbStr != "" {
		database = dbStr
	}
	dbURI := os.Getenv("ASCII_ARCADE_DB_URI")
	if dbURI != "" {
		databaseURI = dbURI
	}

	timeoutStr := os.Getenv("ASCII_ARCADE_PLAYER_TIMEOUT_MINUTES")
	if timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			playerTimeoutMinutes = timeout
		}
	}
}

func GetDebug() bool {
	return debug
}
func GetLogCli() bool {
	return logCli
}
func GetServerHost() string {
	return serverHost
}
func GetServerPortSSH() string {
	return serverPortSSH
}
func GetServerPortWeb() string {
	return serverPortWeb
}
func GetWebAdminKey() string {
	return webAdminKey
}
func GetDatabase() string {
	return database
}
func GetDatabaseURI() string {
	return databaseURI
}
func GetWebAllowedOrigins() string {
	return webAllowedOrigins
}

func GetPlayerTimeoutDuration() time.Duration {
	return time.Duration(playerTimeoutMinutes) * time.Minute
}
