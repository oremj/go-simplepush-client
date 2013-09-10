package main

import (
	"github.com/oremj/go-simplepush-client/pushclient"
)

type Client struct {
	PingSent     int
	PingRecv     int
	Notification chan *pushclient.Notification
	Register     chan *pushclient.RegisterResponse
}

func NewClient() *Client {
	c := &Client{
		Notification: make(chan *pushclient.Notification),
		Register:     make(chan *pushclient.RegisterResponse),
	}
	return c
}

func (c *Client) NotificationHandler(resp *pushclient.Notification) {
	c.Notification <- resp
}

func (c *Client) RegisterHandler(resp *pushclient.RegisterResponse) {
	c.Register <- resp
}
