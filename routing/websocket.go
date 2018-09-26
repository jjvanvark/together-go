package routing

import (
	"encoding/json"
	"errors"
	"log"
	"maus/together-go/database"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) * 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(req *http.Request) bool {
		return true
	},
}

// Hub

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	defer func() {
		log.Println("Exiting hub run")
	}()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Client

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	authorized bool
}

func (c *Client) readPump() {
	defer func() {
		log.Println("defer writepump")
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			log.Println(err)
			break
		}

		err = handleMessage(c, message)
		if err != nil {
			log.Println("Hellooooooo")
			log.Println(err)
		}

	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	select {
	case message, ok := <-c.send:
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println(err)
			return
		}
	case <-ticker.C:
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}

}

func handleMessage(c *Client, message []byte) error {
	token, load, err := parseMessage(message)
	if err != nil {
		return err
	}

	user, err := isSecure(token)
	if err != nil {
		handleUnauthorized(c)
		return err
	}

	switch load {
	case "start":
		err = handleStartMessage(user, c)
		if err != nil {
			return err
		}
	default:
		return errors.New("Unkown load value: " + load)
	}

	return nil
}

func handleUnauthorized(c *Client) error {
	result, err := json.Marshal(struct {
		Status string `json:"status"`
	}{
		"unauthorized",
	})

	if err != nil {
		return err
	}

	c.send <- result
	return nil
}

func handleInternalServerError(c *Client) error {
	result, err := json.Marshal(struct {
		Status string `json:"status"`
	}{
		"server_error",
	})

	if err != nil {
		return err
	}

	c.send <- result
	return nil
}

func handleStartMessage(user *database.User, c *Client) error {
	collaborators, err := db.GetCollaborators(user.Collaborators)
	if err != nil {
		return err
	}

	collaborating, err := db.GetCollaborating(user.Id)
	if err != nil {
		return err
	}

	result, err := json.Marshal(struct {
		Status        string          `json:"status"`
		Load          string          `json:"load"`
		User          *database.User  `json:"user"`
		Collaborators []database.User `json:"collaborators"`
		Collaborating []database.User `json:"collaborating"`
	}{
		"success",
		"start",
		user,
		collaborators,
		collaborating,
	})
	if err != nil {
		return err
	}

	c.send <- result
	return nil
}

func parseMessage(message []byte) (string, string, error) {
	result := struct {
		Token string `json:"token"`
		Load  string `json:"load"`
	}{}

	err := json.Unmarshal(message, &result)
	if err != nil {
		return "", "", err
	}

	return result.Token, result.Load, nil
}

func isSecure(token string) (*database.User, error) {
	id, err := parseToken(token)
	if err != nil {
		return nil, err
	}

	user, err := db.GetUser(*id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func handleWs(hub *Hub, rw http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte),
		authorized: false,
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}
