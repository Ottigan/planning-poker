package types

import "github.com/gofiber/contrib/websocket"

type User struct {
	ID         string
	Name       string
	Active     bool
	Vote       int
	Connection *websocket.Conn
}

type Stats struct {
	Min string
	Avg string
	Max string
}
