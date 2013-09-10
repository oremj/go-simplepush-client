package main

import (
	"github.com/oremj/go-simplepush-client/pushclient"
)

type Client struct {
	PingSent int
	PingRecv int
	Notification chan *pushclient.Notification
	Register chan *pushclient.RegisterResponse
}

func (c *Client) NotificationHandler(resp *pushclient.Notification) {
	c.Notification <- resp
}

func (c *Client) RegisterHandler(resp *pushclient.RegisterResponse) {
	c.Register <- resp
}
