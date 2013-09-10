package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/oremj/go-simplepush-client/pushclient"
	"net/http"
	"strings"
)

var server = flag.String("server", "localhost", "Pushgo Server Name")

type endPoint struct {
	reg *pushclient.RegisterResponse
	version int
}

func SendPing(endPoint string, version int) (err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("PUT", endPoint, strings.NewReader(fmt.Sprintf("version=%d", version)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return
	}
	_, err = client.Do(req)
	return
}

func RunClient(server string) {
	c := &Client{
		Notification: make(chan *pushclient.Notification),
		Register: make(chan *pushclient.RegisterResponse),
	}
	pc, err := pushclient.NewClient(server, 443, true, c)
	if err != nil {
		return
	}
	pc.Register()
	go pc.Run()
	endPoints := make(map[string]*endPoint)
	for {
		select {
		case reg := <-c.Register:
			endPoints[reg.ChannelID] = &endPoint{reg, 1}
			SendPing(reg.PushEndpoint, 1)
			c.PingSent++
		case notif := <-c.Notification:
			for _, update := range notif.Updates {
				e, ok := endPoints[update.ChannelID]
				if ok {
					e.version++
					SendPing(e.reg.PushEndpoint, e.version)
				}
				c.PingRecv++
			}
		}
	}
}

func main() {
	flag.Parse()
	RunClient(*server)
}
