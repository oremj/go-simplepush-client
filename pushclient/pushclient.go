package pushclient

import (
	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/go.net/websocket"
	"crypto/tls"
	"fmt"
	"log"
	"os"
)

type Client struct {
	Ws         *websocket.Conn
	Uaid       string
	ChannelIDs []string
	handler    PushHandler
}

func debug(args ...interface{}) {
	d := os.Getenv("DEBUG")
	if d == "1" {
		log.Println(args...)
	}
}

func NewClient(server string, port int, secure bool, handler PushHandler) (c *Client, err error) {
	var url, origin string
	if secure {
		origin = fmt.Sprintf("https://%s/", server)
		url = fmt.Sprintf("wss://%s:%d/", server, port)
	} else {
		origin = fmt.Sprintf("http://%s/", server)
		url = fmt.Sprintf("ws://%s:%d/", server, port)
	}

	config, err := websocket.NewConfig(url, origin)
	if err != nil {
		return
	}
	if secure {
		config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {
		return
	}

	c = &Client{Ws: ws, Uaid: "", ChannelIDs: make([]string, 0)}
	c.handler = handler
	return
}

func (c *Client) Run() error {
	for {
		resp, err := c.Receive()
		if err != nil {
			debug(err)
			return err
		}
		switch resp["messageType"].(string) {
		case "register":
			go c.handleRegister(resp)
		case "notification":
			go c.handleNotification(resp)
		}
	}
}

func (c *Client) handleRegister(resp Response) {
	register := &RegisterResponse{
		ChannelID:    resp["channelID"].(string),
		Status:       int(resp["status"].(float64)),
		PushEndpoint: resp["pushEndpoint"].(string),
	}
	c.handler.RegisterHandler(register)
}

func (c *Client) handleNotification(resp Response) {
	n := &Notification{}
	for _, up := range resp["updates"].([]interface{}) {
		update := up.(map[string]interface{})
		u := &Update{ChannelID: update["channelID"].(string),
			Version: int(update["version"].(float64))}
		n.Updates = append(n.Updates, u)
	}
	c.handler.NotificationHandler(n)
	c.Send(&AckMessage{MessageType: "ack", Updates: n.Updates})
}

func (c *Client) Send(msg interface{}) (err error) {
	debug("client.send:", msg)
	err = websocket.JSON.Send(c.Ws, msg)
	return
}

func (c *Client) Receive() (resp Response, err error) {
	err = websocket.JSON.Receive(c.Ws, &resp)
	debug("client.recv:", resp)
	return
}

func (c *Client) Handshake() (err error) {
	msg := HandshakeMessage{MessageType: "hello", Uaid: c.Uaid, ChannelIDs: c.ChannelIDs}
	err = c.Send(msg)
	if err != nil {
		return
	}

	resp, err := c.Receive()
	v, ok := resp["uaid"].(string)
	if !ok {
		return fmt.Errorf("uaid in handshake not string")
	}
	c.Uaid = v
	return
}

func (c *Client) Register() (err error) {
	msg := RegisterMessage{MessageType: "register", ChannelID: uuid.New()}
	err = c.Send(msg)
	return
}
