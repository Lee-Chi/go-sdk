package ws

import (
	"fmt"
	"net/http"
)

type Hub struct {
	connections map[ID]*Connection

	broadcast chan Packet

	relay chan Packet

	register chan *Connection

	unregister chan *Connection

	log chan string

	logger func(log string)

	handlers map[string]Handler

	running bool
}

func NewHub(logger func(log string)) *Hub {
	return &Hub{
		broadcast:   make(chan Packet, 256),
		relay:       make(chan Packet, 256),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[ID]*Connection),
		log:         make(chan string, 256),
		logger:      logger,
		handlers:    make(map[string]Handler),
		running:     false,
	}
}

func (h *Hub) RegisterHandler(name string, handler Handler) {
	if h.running {
		h.log <- "hub is running, cannot relay message"
		return
	}

	h.handlers[name] = handler
}

func (hub *Hub) Accept(w http.ResponseWriter, r *http.Request) (ID, error) {
	// Upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return "", fmt.Errorf("could not upgrade connection to websocket: %w", err)
	}

	id := NewID()

	// Register our new connection
	connection := NewConnection(id, hub, ws)

	hub.register <- connection

	connection.daemon()

	return id, nil
}

func (h *Hub) Broadcast(cmd Command) {
	h.broadcast <- Packet{Message: cmd.Marshal()}
}

func (h *Hub) Relay(to ID, cmd Command) {
	h.relay <- Packet{To: to, Message: cmd.Marshal()}
}

func (h *Hub) Run(shutdown chan bool) {
	h.running = true

	stop := make(chan bool)
	go h.dump(stop)
	defer func() {
		stop <- true
		h.running = false
	}()

	for {
		select {
		case <-shutdown:
			close(h.register)
			for _, connection := range h.connections {
				connection.socket.Close()
			}
			return
		case connection := <-h.register:
			h.connections[connection.id] = connection
		case connection := <-h.unregister:
			if _, ok := h.connections[connection.id]; ok {
				delete(h.connections, connection.id)
				close(connection.send)
			}
		case packet := <-h.broadcast:
			for _, connection := range h.connections {
				select {
				case connection.send <- packet.Message:
				default:
					close(connection.send)
					delete(h.connections, connection.id)
				}
			}
		case packet := <-h.relay:
			if connection, ok := h.connections[packet.To]; ok {
				select {
				case connection.send <- packet.Message:
				default:
					close(connection.send)
					delete(h.connections, connection.id)
				}
			}
		}
	}
}

func (h *Hub) dump(stop chan bool) {
	for {
		select {
		case log := <-h.log:
			if h.logger != nil {
				h.logger(log)
			}
		case <-stop:
			return
		}
	}
}
