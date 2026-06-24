package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	busID := c.Param("id")
	if busID == "" {
		c.JSON(400, gin.H{"error": "bus id required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		Conn:  conn,
		Send:  make(chan BroadcastMessage, 16),
		BusID: busID,
		Hub:   h.hub,
	}

	h.hub.register <- client

	go client.ReadPump()
	go client.WritePump()
}
