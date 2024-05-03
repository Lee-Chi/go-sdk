package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	id     ID
	hub    *Hub
	socket *websocket.Conn
	send   chan []byte
	closes map[string]Handler
	mtx    sync.Mutex
}

func NewConnection(id ID, hub *Hub, socket *websocket.Conn) *Connection {
	return &Connection{
		id:     id,
		hub:    hub,
		socket: socket,
		send:   make(chan []byte, 32),

		closes: make(map[string]Handler),
		mtx:    sync.Mutex{},
	}
}

func (c *Connection) ID() ID {
	return c.id
}

func (c *Connection) Log(log string) {
	c.hub.log <- fmt.Sprintf("connection %s, %s", c.id, log)
}

func (c *Connection) Send(to ID, cmd *Command) {
	if to == c.ID() {
		c.send <- cmd.Marshal()
		return
	}

	c.hub.relay <- Packet{To: to, Message: cmd.Marshal()}
}

func (c *Connection) Broadcast(cmd *Command) {
	c.hub.broadcast <- Packet{Message: cmd.Marshal()}
}

func (c *Connection) RegisterCloseHandler(name string, handler Handler) {
	c.mtx.Lock()
	c.closes[name] = handler
	c.mtx.Unlock()
}

func (c *Connection) UnregisterCloseHandler(name string) {
	c.mtx.Lock()
	delete(c.closes, name)
	c.mtx.Unlock()
}

func (c *Connection) daemon() {
	go c.read()
	go c.write()
}

func (c *Connection) read() {
	defer func() {
		c.hub.log <- fmt.Sprintf("connection %s, leave read", c.id)
		c.hub.unregister <- c

		c.mtx.Lock()
		for _, handler := range c.closes {
			handler(c, nil)
		}
		c.mtx.Unlock()

		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log <- fmt.Sprintf("connection %s, failed to read message: %s", c.id, err)
			} else {
				c.hub.log <- fmt.Sprintf("connection %s, close message: %s", c.id, err)
			}
			break
		}

		var cmd Command
		if err := cmd.Unmarshal(message); err != nil {
			c.hub.log <- fmt.Sprintf("connection %s receive(%s), failed to unmarshal message: %s", c.id, string(message), err)
			continue
		}

		handler, ok := c.hub.handlers[cmd.Name]
		if !ok {
			c.hub.log <- fmt.Sprintf("connection %s recieve(%s), unknown API: %s", c.id, string(message), cmd.Name)
			continue
		}

		go handler(c, &cmd)
	}
}

func (c *Connection) write() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		c.hub.log <- fmt.Sprintf("connection %s, leave write", c.id)
		ticker.Stop()
		c.socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.socket.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.hub.log <- fmt.Sprintf("connection %s, failed to get message", c.id)
				// The hub closed the channel.
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.socket.WriteMessage(websocket.TextMessage, message); err != nil {
				c.hub.log <- fmt.Sprintf("connection %s, failed to write message: %s", c.id, err)
				return
			}
		case <-ticker.C:
			c.socket.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.log <- fmt.Sprintf("connection %s, failed to write ping message: %s", c.id, err)
				return
			}
		}
	}
}
