package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/oremj/go-simplepush-client/pushclient"
	"log"
	"net/http"
	"strings"
)

var server = flag.String("server", "localhost", "Pushgo Server Name")
var secure = flag.Bool("secure", false, "Use wss/https")
var numClients = flag.Int("clients", 1, "Number of concurrent clients")

type endPoint struct {
	reg     *pushclient.RegisterResponse
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
	res, err := client.Do(req)
	if err == nil {
		res.Body.Close()
	}
	return
}

func RunClient(server string, port int, secure bool, c *Client) {
	pc, err := pushclient.NewClient(server, port, secure, c)
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
			err := SendPing(reg.PushEndpoint, 1)
			if err != nil {
				statChan <- &stat{"put_fail", 1}
			} else {
				statChan <- &stat{"ping_sent", 1}
			}
		case notif := <-c.Notification:
			for _, update := range notif.Updates {
				e, ok := endPoints[update.ChannelID]
				if ok {
					e.version++
					err := SendPing(e.reg.PushEndpoint, e.version)
					if err != nil {
						log.Println(err)
						statChan <- &stat{"put_fail", 1}
					} else {
						statChan <- &stat{"ping_sent", 1}
					}
				}
				statChan <- &stat{"ping_recv", 1}
			}
		}
	}
}

type stat struct {
	key string
	val int
}

var statChan chan *stat

func main() {
	statChan = make(chan *stat, 100000)
	metrics := make(map[string]int)
	flag.Parse()
	port := 80
	if *secure {
		port = 443
	}
	clients := make([]*Client, 0, *numClients)
	for i := 0; i < *numClients; i++ {
		c := NewClient()
		clients = append(clients, c)
		go RunClient(*server, port, *secure, c)
	}
	for {
		s := <-statChan
		metrics[s.key] += s.val
		fmt.Println(metrics)
	}
}
