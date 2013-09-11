package main

import (
	"flag"
	"fmt"
	"github.com/oremj/go-simplepush-client/pushclient"
	"time"
)

var server = flag.String("server", "localhost", "Pushgo Server Name")
var secure = flag.Bool("secure", false, "Use wss/https")
var numClients = flag.Int("clients", 1, "Number of concurrent clients")

func RunClient(server string, port int, secure bool, c *Client, wait bool) {
	pc, err := pushclient.NewClient(server, port, secure, c)
	if err != nil {
		counterChan <- &stat{"conn_fail", 1}
		time.Sleep(5 * time.Second)
		go RunClient(server, port, secure, c, false)
		return
	}
	counterChan <- &stat{"conn_ok", 1}
	disconnected := make(chan bool)
	go func(c *pushclient.Client) {
		err := c.Run()
		if err != nil {
			disconnected <- true
		}
	}(pc)
	if wait {
		<-waitConnect
	}
	pc.Register()

	endPoints := make(map[string]*endPoint)
	for {
		select {
		case <-disconnected:
			counterChan <- &stat{"conn_lost", 1}
			time.Sleep(5 * time.Second)
			go RunClient(server, port, secure, c, false)
			return
		case reg := <-c.Register:
			e := NewEndpoint(reg)
			endPoints[reg.ChannelID] = e
			e.sendPing()
		case notif := <-c.Notification:
			for _, update := range notif.Updates {
				e, ok := endPoints[update.ChannelID]
				if ok {
					select {
					case e.done <- true:
					default:
					}
					e.version++
					e.sendPing()
				}
			}
		}
	}

}

type stat struct {
	key string
	val int
}

var counterChan chan *stat
var waitConnect chan bool

func main() {
	flag.Parse()
	counterChan = make(chan *stat, 100000)
	waitConnect = make(chan bool, *numClients)
	metrics := make(map[string]int)

	port := 80
	if *secure {
		port = 443
	}
	clients := make([]*Client, 0, *numClients)
	for i := 0; i < *numClients; i++ {
		c := NewClient()
		clients = append(clients, c)
		go RunClient(*server, port, *secure, c, true)
		time.Sleep(100 * time.Millisecond)
	}
	connected := false
	for {
		s := <-counterChan
		metrics[s.key] += s.val
		if !connected && metrics["conn_ok"]+metrics["conn_fail"] >= *numClients {
			for i := 0; i < *numClients; i++ {
				waitConnect <- true
			}
			connected = true
		}
		fmt.Println(metrics)
	}
}
