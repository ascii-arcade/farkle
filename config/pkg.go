package config

var (
	debug      bool
	serverURL  string = "localhost"
	serverPort string = "8080"
)

func GetServerURL() string {
	return serverURL
}
func GetServerPort() string {
	return serverPort
}
func GetDebug() bool {
	return debug
}

func SetServerURL(url *string) {
	serverURL = *url
}
func SetServerPort(port *string) {
	serverPort = *port
}
func SetDebug(d *bool) {
	debug = *d
}
