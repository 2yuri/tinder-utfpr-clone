package websocket

import (
	"context"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type Hub struct {
	mu sync.Mutex
	//map[userId]clients
	subs       map[string]map[*client]struct{}
	unregister chan *client
}

//NewHub create a hub to manage new websocket connection
func NewHub() *Hub {
	return &Hub{
		unregister: make(chan *client),
		subs:       make(map[string]map[*client]struct{}),
	}
}

//StartServer start the hub to receive clients and send messages
func (h *Hub) StartServer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("delete all client and close gracefully")
			h.deleteAll()
			return
		case client := <-h.unregister:
			h.handleClientDelete(client)
		case event := <-Events:
			for c := range h.subs[event.UserID] {
				if c.IsAlive() {
					c.out <- event
				}
			}
		}
	}
}

//RemoveClient remove client from hub
func (h *Hub) RemoveClient(c *client) {
	h.unregister <- c
}

func (h *Hub) HandleClientInsertion(c *client, userId string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subs[userId]; !ok {
		h.subs[userId] = make(map[*client]struct{})
	}

	value := h.subs[userId]
	value[c] = struct{}{}

	h.subs[userId] = value
}

func (h *Hub) deleteAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, clients := range h.subs {
		for cl := range clients {
			cl.close()
			delete(clients, cl)
		}
	}

	return
}

func (h *Hub) handleClientDelete(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	c.close()
	for _, clients := range h.subs {
		for cl := range clients {
			if cl == c {
				delete(clients, c)
			}
		}
	}
}

func Subscribe(r *gin.RouterGroup, hhh *Hub, authMiddleware gin.HandlerFunc) {
	r.GET("/subscribe", authMiddleware, func(c *gin.Context) {
		upgrader := ws.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		id := c.GetString("userId")
		if id == "" {
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("cannot upgrade: ", err.Error())
			return
		}

		cl := NewClient(conn, hhh)
		hhh.HandleClientInsertion(cl, id)
	})
}
