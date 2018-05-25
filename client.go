package simpleWebsocket

import (
	"time"
	"github.com/gorilla/websocket"
	"log"
	"bytes"
	"encoding/json"
	"fmt"
)

type Client struct {
	server *Server
	conn   *websocket.Conn
	send   chan []byte
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	readLimit    int64 = 512
	readDeadLine       = time.Duration(60)
)

func (c *Client) Send(message interface{}) {

	var sendBytes []byte
	var ok bool

	sendBytes, err := json.Marshal(message)

	if err != nil {
		if sendBytes, ok = message.([]byte); ok {
			fmt.Println("can't send message, bad message type")
		}

	}

	if len(sendBytes) == 0 {
		fmt.Println("empty send data")
		return
	}

	c.send <- sendBytes
}

func (c *Client) SendString(message string) {
	c.send <- []byte(message)
}

func (c *Client) read() {

	defer c.conn.Close()
	c.conn.SetReadLimit(readLimit)
	c.conn.SetReadDeadline(time.Now().Add(readDeadLine * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(readDeadLine * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}

			c.server.onDisconnect(c)
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.server.handleMessage(c, message)

	}
}

func (c *Client) write() {

	ticker := time.NewTicker(15 * time.Second)

	defer close(c.send)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

			break

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}

}
