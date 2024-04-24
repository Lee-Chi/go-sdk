package ws

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// Maximum message size allowed from peer.
	MaxMessageSize = 512
)

type Client struct {
	id      string
	hub     *Hub
	socket  *websocket.Conn
	send    chan []byte
	destroy func(*Client)
}

func NewClient(id string, hub *Hub, socket *websocket.Conn, destroy func(*Client)) *Client {
	return &Client{
		id:      id,
		hub:     hub,
		socket:  socket,
		send:    make(chan []byte, 256),
		destroy: destroy,
	}
}

func (c Client) ID() string {
	return c.id
}

func (c Client) Hub() *Hub {
	return c.hub
}

func (c *Client) Run() {
	go c.read()
	go c.write()
}

type Command struct {
	API     string `json:"api"`
	Message string `json:"message"`
}

func (c *Client) read() {
	defer func() {
		c.hub.log <- fmt.Sprintf("client %s, leave read", c.id)
		c.hub.unregister <- c
		c.destroy(c)
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log <- fmt.Sprintf("client %s, failed to read message: %s", c.id, err)
			} else {
				c.hub.log <- fmt.Sprintf("client %s, close message: %s", c.id, err)
			}
			break
		}

		var command Command
		if err := json.Unmarshal(message, &command); err != nil {
			c.hub.log <- fmt.Sprintf("client %s receive(%s), failed to unmarshal message: %s", c.id, string(message), err)
			continue
		}

		handler, ok := c.hub.handlers[command.API]
		if !ok {
			c.hub.log <- fmt.Sprintf("client %s recieve(%s), unknown API: %s", c.id, string(message), command.API)
			continue
		}

		go handler(c, command.Message)
	}
}

func (c *Client) write() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		c.hub.log <- fmt.Sprintf("client %s, leave write", c.id)
		ticker.Stop()
		c.socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.socket.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.hub.log <- fmt.Sprintf("client %s, failed to get message", c.id)
				// The hub closed the channel.
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.socket.NextWriter(websocket.TextMessage)
			if err != nil {
				c.hub.log <- fmt.Sprintf("client %s, failed to get next writer: %s", c.id, err)
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				c.hub.log <- fmt.Sprintf("client %s, failed to close writer: %s", c.id, err)
				return
			}
		case <-ticker.C:
			c.socket.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.log <- fmt.Sprintf("client %s, failed to write ping message: %s", c.id, err)
				return
			}
		}
	}
}
