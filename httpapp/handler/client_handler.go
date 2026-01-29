package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// When a message is stored it is send to the MessageStore via a channel

type MessageStore struct {
	MsgChan chan Message // CSP channel for incoming messages
	// Todo: Add client conections to the MessageStore
	// When the MessageStore receives a message it will write to the all clients channel
	Clients []*ClientConnection
}

// ClientConnection represents a WebSocket connection (actor)
type ClientConnection struct {
	conn *websocket.Conn
	send chan []byte // Channel for outgoing messages

}

func broadCastToRegisteredClients() {

	for message := range MessageStoreInstance.MsgChan {
		for _, client := range MessageStoreInstance.Clients {
			log.Printf("Broadcasting message to client: %+v", message)
			client.send <- []byte(message.Message)

		}
	}
}

func (c *ClientConnection) writeToSocket() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, ([]byte(message)))
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}
func (c *ClientConnection) writeToSocketOld() {
	defer func() {
		c.conn.Close()
	}()

	for message := range MessageStoreInstance.MsgChan {
		err := c.conn.WriteMessage(websocket.TextMessage, ([]byte(message.Message)))
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}
func (c *ClientConnection) Start() {
	go c.writeToSocket()
}
func WsClientHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	client := &ClientConnection{
		conn: conn,
		send: make(chan []byte, 256),
	}
	MessageStoreInstance.Clients = append(MessageStoreInstance.Clients, client)
	log.Println("New WebSocket connection established")
	client.Start()
}
