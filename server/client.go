package server

import (
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/dobegor/chat/common"
	"github.com/gorilla/websocket"
)

// Represents a connected client.
type Client struct {
	// Client's username
	Username string

	// Connection pool
	conns []*connection

	// Synchronizes access to connection pool
	m sync.Mutex

	hub *Hub

	Online bool
}

type connection struct {
	username string
	*websocket.Conn
}

func (c *Client) removeConnectionByIndex(i int) {
	go log.WithFields(log.Fields{
		"ip":   c.conns[i].RemoteAddr().String(),
		"user": c.conns[i].username,
		"time": time.Now().String(),
	}).Info("Removing connection")

	// we don't actually care about active connections order,
	// so let's just swap deleted item with the last one and shrink the slice
	c.conns[i] = c.conns[len(c.conns)-1]
	c.conns = c.conns[:len(c.conns)-1]
}

// Removes connection from pool.
// Returns true if there are connections left.
func (c *Client) RemoveConnection(conn *connection) bool {
	c.m.Lock()
	defer c.m.Unlock()

	for i, cnn := range c.conns {
		if cnn == conn {
			c.removeConnectionByIndex(i)
			return len(c.conns) > 0
		}
	}

	return false
}

func (c *Client) Connections() int {
	return len(c.conns)
}

// Adds a websocket connection to connection pool.
func (c *Client) AddConnection(conn *connection) {
	c.m.Lock()
	defer c.m.Unlock()

	go log.WithFields(log.Fields{
		"ip":   conn.RemoteAddr().String(),
		"user": c.Username,
		"time": time.Now().String(),
	}).Info("Adding connection")

	c.conns = append(c.conns, conn)
}

func (c *Client) Send(msg *common.Message) {
	c.m.Lock()
	defer c.m.Unlock()

	toRemove := []int{}

	// send message to all the connections which are associated with this user
	for i, conn := range c.conns {
		err := conn.WriteMessage(websocket.BinaryMessage, common.Encode(msg))

		if err != nil {
			// this connection is broken, mark it for removal
			toRemove = append(toRemove, i)
		}
	}

	// remove all broken connections
	for i := range toRemove {
		c.removeConnectionByIndex(i)
	}
}
