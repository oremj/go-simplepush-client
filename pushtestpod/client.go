package main

import (
	"github.com/oremj/go-simplepush-client/pushclient"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Client struct {
	server       string
	port         int
	Notification chan *pushclient.Notification
	RegisterChan chan *pushclient.RegisterResponse
	config       *Config
	*pushclient.Client
}

type Config struct {
	secure bool
	delay  int
}

func NewClient(server string, port int, config *Config) *Client {
	c := &Client{
		server:       server,
		port:         port,
		config:       config,
		Notification: make(chan *pushclient.Notification),
		RegisterChan: make(chan *pushclient.RegisterResponse),
	}
	return c
}

func (c *Client) Connect() (err error) {
	c.Client, err = pushclient.NewClient(c.server, c.port, c.config.secure, c)
	if err != nil {
		incStat("conn_fail")
		return
	}
	incStat("conn_ok")
	return
}

func (c *Client) Handshake() (err error) {
	err = c.Client.Handshake()
	if err != nil {
		incStat("handshake_fail")
		return
	}
	incStat("handshake_ok")
	return
}

func (c *Client) Run() (err error) {
	disconnected := make(chan error)
	go func() {
		err := c.Client.Run()
		if err != nil {
			disconnected <- err
		}
	}()

	endPoints := make(map[string]*endPoint)
	for {
		select {
		case err = <-disconnected:
			incStat("conn_lost")
			return
		case reg := <-c.RegisterChan:
			incStat("reg_ok")
			e := NewEndpoint(reg)
			endPoints[reg.ChannelID] = e
			go func() {
				e.run(c.config.delay)
				c.Register()
			}()
		case notif := <-c.Notification:
			for _, update := range notif.Updates {
				e, ok := endPoints[update.ChannelID]
				if ok {
					select {
					case e.notify <- true:
					default:
					}
				}
			}
		}
	}
}

func (c *Client) Register() {
	incStat("reg_try")
	c.Client.Register()
}

func (c *Client) NotificationHandler(resp *pushclient.Notification) {
	c.Notification <- resp
}

func (c *Client) RegisterHandler(resp *pushclient.RegisterResponse) {
	c.RegisterChan <- resp
}
