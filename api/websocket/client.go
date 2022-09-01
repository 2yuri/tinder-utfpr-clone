package websocket

import (
	"context"
	ws "github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type client struct {
	aliveLock sync.Mutex
	alive     bool
	hub       *Hub
	conn      *ws.Conn
	out       chan interface{}
	ctx       context.Context
	cancel    context.CancelFunc
}

//NewClient create a new client, add into hub and start ping and watch
func NewClient(conn *ws.Conn, hub *Hub) *client {
	client := &client{conn: conn, hub: hub, out: make(chan interface{}, outChannelSize), alive: true}
	client.ctx, client.cancel = context.WithCancel(context.Background())
	go client.loopIn()
	go client.loopOut()

	return client
}

func (c *client) IsAlive() bool {
	c.aliveLock.Lock()
	defer c.aliveLock.Unlock()
	return c.alive
}

//close call this function to close client connection and remove from hub
func (c *client) close() {
	c.aliveLock.Lock()
	defer c.aliveLock.Unlock()
	if c.alive {
		c.conn.Close()
		c.alive = false
		c.cancel()
		close(c.out)
		for len(c.out) > 0 {
			<-c.out
		}
	}
}

//watch this function read client messages to check if client cancel the conn or send a ping
func (c *client) loopIn() {
	defer func() {
		c.close()
		c.hub.handleClientDelete(c)
	}()

	for {
		messageType, _, err := c.conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Println("ws.loopIn ", "err ", err.Error())
			}
			break
		}

		switch messageType {
		case ws.CloseMessage:
			break
		case ws.PingMessage:
			c.conn.WriteControl(ws.PongMessage, nil, time.Now().Add(pingPeriod))
		}
	}
}

func (c *client) loopOut() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case m := <-c.out:
			err := c.conn.WriteJSON(m)
			if err != nil {
				log.Println("ws.loopOut", "err", err.Error())
				c.close()
			}
		}
	}
}
