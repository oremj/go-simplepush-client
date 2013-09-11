package main

import (
	"flag"
	"fmt"
	"time"
)

var server = flag.String("server", "localhost", "Pushgo Server Name")
var secure = flag.Bool("secure", false, "Use wss/https")
var numClients = flag.Int("clients", 1, "Number of concurrent clients")
var delay = flag.Int("delay", 10000, "Delay between PUTs in milliseconds.")

func incStat(name string) {
	counterChan <- &stat{name, 1}
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
		c := NewClient(*server, port, &Config{secure: *secure, delay: *delay})
		clients = append(clients, c)
		go func(c *Client) {
			for {
				err := c.Run()
				if err != nil {
					time.Sleep(2 * time.Second)
				}
			}
		}(c)
	}

	go func() {
		connected := 0
		for s := range counterChan {
			metrics[s.key] += s.val
			switch s.key {
			case "conn_ok":
				connected++
				if connected == *numClients {
					for _, c := range clients {
						c.pc.Register()
					}
				}
			}
		}
	}()

	for {
		fmt.Println(metrics)
		time.Sleep(1 * time.Second)
	}
}
