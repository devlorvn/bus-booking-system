package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn  *websocket.Conn
	Send  chan BroadcastMessage
	Hub   *Hub
	BusID string
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			break
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		select {
		case c.Hub.unregister <- c:
		case <-time.After(100 * time.Millisecond):
		}
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
