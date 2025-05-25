package config

var (
	debug      bool
	serverURL  string
	serverPort string
	secure     bool
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
func GetSecure() bool {
	return secure
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
func SetSecure(s *bool) {
	secure = *s
}
