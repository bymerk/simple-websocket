package main

import (
	"simple-websocket"
	"fmt"
)

func main()  {

	ws := simpleWebsocket.New()

	ws.OnConnect(func(c *simpleWebsocket.Client) {
		c.SendString("hello, bro")
	})

	// {"method": "method_name": "params": {"some": "params"}}
	ws.On("echo", func(c *simpleWebsocket.Client, message simpleWebsocket.Message) {
		c.Send(message)
	})

	ws.OnBytes(func(msg []byte, c *simpleWebsocket.Client) {
		fmt.Println(string(msg))
	})

	ws.OnDisconnect(func(c *simpleWebsocket.Client) {
		c.SendString("see you later")
	})

	ws.Run(":8080")
}
