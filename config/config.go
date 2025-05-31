package config

import "os"

var (
	debug  bool
	logCli bool

	serverHost    string
	serverPortSSH string
	serverPortWeb string
	webAdminKey   string
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
