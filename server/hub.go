package server

import (
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/dobegor/chat/common"
	"github.com/gorilla/websocket"
)

// Hub receives and broadcast messages from each user.
// Hub also broadcasts if someone goes online or offline.
type Hub struct {
	clients map[string]*Client
	m       sync.Mutex
}

func (h *Hub) Init() {
	h.clients = map[string]*Client{}
}

// Adds connection to Hub.
// Either registers a new Client or adds connection to existing one's connection pool.
func (h *Hub) AddConnection(username string, conn *websocket.Conn) {
	h.m.Lock()
	defer h.m.Unlock()

	c := &connection{username, conn}

	if user, ok := h.clients[c.username]; ok {
		// we already have this user, let's add this connection to his pool
		user.AddConnection(c)

		// launch goroutine which reads messages from client
		go h.readPump(c)

		if !h.clients[c.username].Online {
			h.clients[c.username].Online = true
			h.broadcast(&common.Message{
				Username: "SYSTEM",
				Content:  "User online: " + c.username,
			})
		}
		return
	}

	go log.WithFields(log.Fields{
		"username": c.username,
		"time":     time.Now().String(),
	}).Info("Registering new client")

	client := &Client{
		Username: c.username,
		conns:    []*connection{c},
		hub:      h,
	}

	h.clients[c.username] = client

	// launch goroutine which reads messages from client
	go h.readPump(c)

	h.broadcast(&common.Message{
		Username: "SYSTEM",
		Content:  "User online: " + c.username,
	})

}

func (h *Hub) readPump(c *connection) {
	for {
		_, rawMsg, err := c.ReadMessage()

		if err != nil {
			go log.WithFields(log.Fields{
				"user": c.username,
				"err":  err.Error(),
				"time": time.Now().String(),
			}).Error("Error reading connection")

			c.Close()

			h.m.Lock()
			defer h.m.Unlock()

			if !h.clients[c.username].RemoveConnection(c) {
				h.broadcast(&common.Message{
					Username: "SYSTEM",
					Content:  "User went offline: " + c.username,
				})

				h.clients[c.username].Online = false
			}
			break
		}

		msg, err := common.Decode(rawMsg)

		if err != nil {
			go log.WithFields(log.Fields{
				"user": c.username,
				"err":  err.Error(),
				"time": time.Now().String(),
			}).Error("Error decoding message")

			// can't decode, so we can't broadcast this
			continue
		}

		msg.Username = c.username
		h.Broadcast(msg)
	}
}

// Broadcasts message to all connected clients
func (h *Hub) Broadcast(m *common.Message) {
	go log.WithFields(log.Fields{
		"message": m.Content,
		"author":  m.Username,
		"time":    time.Now().String(),
	}).Info("Broadcasting message")

	h.m.Lock()
	defer h.m.Unlock()

	h.broadcast(m)
}

func (h *Hub) broadcast(m *common.Message) {
	for _, client := range h.clients {
		client.Send(m)
	}
}
