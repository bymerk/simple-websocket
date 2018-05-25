package simpleWebsocket

import (
	"net/http"
	"github.com/gorilla/websocket"
	"sync"
	"encoding/json"
	"errors"
)

type Server struct {
	HttpServer   *http.ServeMux
	WsPath       string
	Upgrader     websocket.Upgrader
	methods      *sync.Map
	onConnect    handleClient
	onDisconnect handleClient
	onBytes      handleClientsBytes
}

type Message struct {
	Method string      `json:"method"`
	Params json.RawMessage `json:"params"`
}

type handleClient func(c *Client)
type handleClientMethod func(c *Client, message Message)
type handleClientsBytes func(msg []byte, c *Client)

func New() (*Server) {

	server := &Server{
		HttpServer: http.NewServeMux(),
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		methods:      &sync.Map{},
		WsPath:       "/ws",
		onConnect:    func(c *Client) {},
		onDisconnect: func(c *Client) {},
		onBytes:      func(msg []byte, c *Client) {},
	}

	server.initHandler()
	return server
}

func (s *Server) On(method string, cb handleClientMethod) {
	s.methods.Store(method, cb)
}

func (s *Server) OnBytes(cb handleClientsBytes) {
	s.onBytes = cb
}

func (s *Server) OnConnect(cb handleClient) {
	s.onConnect = cb
}

func (s *Server) OnDisconnect(cb handleClient) {
	s.onDisconnect = cb
}

func (s *Server) handleMessage(c *Client, msg []byte) error {

	var message Message

	err := json.Unmarshal(msg, &message)

	if err != nil {
		s.onBytes(msg, c)
		return nil
	}

	if callback, ok := s.methods.Load(message.Method); ok {

		cb, ok := callback.(handleClientMethod)

		if !ok {
			return errors.New("can't get callback")
		}

		cb(c, message)

	} else {
		return errors.New("can't find method")
	}

	return nil
}

func (s *Server) initHandler() {
	s.HttpServer.HandleFunc(s.WsPath, func(w http.ResponseWriter, r *http.Request) {

		conn, err := s.Upgrader.Upgrade(w, r, nil)

		if err != nil {
			r.Body.Close()
		}

		c := &Client{
			server: s,
			conn:   conn,
			send:   make(chan []byte),
		}

		go c.write()

		s.onConnect(c)
		c.read()
	})
}

func (s *Server) Run(addr string) {

	http.ListenAndServe(addr, s.HttpServer)
}

func (s *Server) RunTLS(addr, crt, key string) {
	s.initHandler()
	http.ListenAndServeTLS(addr, crt, key, s.HttpServer)
}
