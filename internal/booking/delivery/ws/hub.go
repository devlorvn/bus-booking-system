package ws

import "context"

type Hub struct {
	rooms map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan RoomBroadcast
	done       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan RoomBroadcast),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	defer close(h.done)
	for {
		select {
		case <-ctx.Done():
			h.cleanup()
			return
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

func (h *Hub) cleanup() {
	for _, room := range h.rooms {
		for client := range room {
			close(client.Send)
			client.Conn.Close()
		}
	}
	h.rooms = make(map[string]map[*Client]bool)
}

func (h *Hub) handleRegister(client *Client) {
	if _, exists := h.rooms[client.BusID]; !exists {
		h.rooms[client.BusID] = make(map[*Client]bool)
	}
	h.rooms[client.BusID][client] = true
}

func (h *Hub) handleUnregister(client *Client) {
	room, exists := h.rooms[client.BusID]
	if !exists {
		return
	}

	if _, ok := room[client]; ok {
		delete(room, client)
		close(client.Send)
	}

	if len(room) == 0 {
		delete(h.rooms, client.BusID)
	}
}

func (h *Hub) handleBroadcast(msg RoomBroadcast) {
	room, exists := h.rooms[msg.BusID]
	if !exists {
		return
	}

	for client := range room {
		select {
		case client.Send <- msg.Message:
		default:
			close(client.Send)
			delete(room, client)
		}
	}
}

func (h *Hub) Broadcast(busID string, msg BroadcastMessage) {
	select {
	case <-h.done:
		return
	case h.broadcast <- RoomBroadcast{
		BusID:   busID,
		Message: msg,
	}:
	}
}
