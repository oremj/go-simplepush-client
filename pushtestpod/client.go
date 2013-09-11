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
	Register     chan *pushclient.RegisterResponse
	config       *Config
	pc           *pushclient.Client
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
		Register:     make(chan *pushclient.RegisterResponse),
	}
	return c
}

func (c *Client) Run() (err error) {
	disconnected := make(chan error)
	c.pc, err = pushclient.NewClient(c.server, c.port, c.config.secure, c)
	if err != nil {
		incStat("conn_fail")
		return
	}
	incStat("conn_ok")
	go func() {
		err := c.pc.Run()
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
		case reg := <-c.Register:
			e := NewEndpoint(reg)
			endPoints[reg.ChannelID] = e
			go func() {
				e.sendPing()
			}()
		case notif := <-c.Notification:
			for _, update := range notif.Updates {
				e, ok := endPoints[update.ChannelID]
				if ok {
					select {
					case e.done <- true:
						incStat("update_ok")
					case <-e.done:
						incStat("update_timeout")
						c.pc.Register()
					}
					e.version++
					go func() {
						delay := c.config.delay * 1000
						range_ := delay / 2
						delay = rand.Intn(range_*2) + (delay - range_)
						time.Sleep(time.Duration(delay) * time.Millisecond)
						e.sendPing()
					}()
				}
			}
		}
	}
}

func (c *Client) NotificationHandler(resp *pushclient.Notification) {
	c.Notification <- resp
}

func (c *Client) RegisterHandler(resp *pushclient.RegisterResponse) {
	c.Register <- resp
}
