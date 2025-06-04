package config

import "os"

const (
	MinimumHeight = 33
	MinimumWidth  = 120
)

var (
	debug  bool = false
	logCli bool = false

	serverHost    string = "localhost"
	serverPortSSH string = "2222"
	serverPortWeb string = "8080"
	webAdminKey   string

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
