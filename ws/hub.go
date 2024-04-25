package ws

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Target  string
	Content []byte
}

type Hub struct {
	clients map[string]*Client

	broadcast chan Message

	relay chan Message

	register chan *Client

	unregister chan *Client

	log chan string

	logger func(log string)

	handlers map[string]func(*Client, string)

	destroyClient func(client *Client)

	running bool
}

func NewHub(logger func(log string)) *Hub {
	return &Hub{
		broadcast:  make(chan Message, 256),
		relay:      make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		log:        make(chan string, 256),
		logger:     logger,
		handlers:   make(map[string]func(*Client, string)),
		running:    false,
	}
}

func (h *Hub) SetDestroyClient(destroy func(client *Client)) {
	h.destroyClient = destroy
}

func (h *Hub) RegisterHandler(api string, handler func(client *Client, message string)) {
	if h.running {
		h.log <- "hub is running, cannot relay message"
		return
	}

	h.handlers[api] = handler
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (hub *Hub) Accept(w http.ResponseWriter, r *http.Request) (string, error) {
	// Upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return "", fmt.Errorf("could not upgrade connection to websocket: %w", err)
	}

	id := uuid.New().String()

	// Register our new client
	client := NewClient(id, hub, ws, hub.destroyClient)

	hub.register <- client

	client.Run()

	return id, nil
}

func (h *Hub) Broadcast(command Command) {
	h.broadcast <- Message{Content: command.Marshal()}
}

func (h *Hub) Relay(target string, command Command) {
	h.relay <- Message{Target: target, Content: command.Marshal()}
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
			return
		case client := <-h.register:
			h.clients[client.id] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
			}
		case message := <-h.broadcast:
			for _, client := range h.clients {
				select {
				case client.send <- message.Content:
				default:
					close(client.send)
					delete(h.clients, client.id)
				}
			}
		case message := <-h.relay:
			if client, ok := h.clients[message.Target]; ok {
				select {
				case client.send <- message.Content:
				default:
					close(client.send)
					delete(h.clients, client.id)
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
