package server

type lobby struct {
	clients map[*client]bool
}
