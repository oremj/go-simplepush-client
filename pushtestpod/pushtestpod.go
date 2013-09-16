package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

var servers = flag.String("servers", "localhost", "Pushgo Servers comma separated.")
var secure = flag.Bool("secure", false, "Use wss/https")
var numClients = flag.Int("clients", 1, "Number of concurrent clients per server")
var delay = flag.Int("delay", 10000, "Delay between PUTs in milliseconds.")
var connectionLimit = flag.Int("connectlimit", 100, "Connection limiter.")

func incStat(name string) {
	counterChan <- &stat{name, 1}
}

type stat struct {
	key string
	val int
}

var counterChan chan *stat

func main() {
	flag.Parse()
	counterChan = make(chan *stat, 100000)
	port := 80
	if *secure {
		port = 443
	}

	metrics := make(Metrics)
	go func() {
		for s := range counterChan {
			metrics[s.key] += s.val
		}
	}()

	serverList := strings.Split(*servers, ",")
	go func() {
		totalClients := *numClients * len(serverList)
		waitHandshake := make(chan bool)
		waitRegister := make(chan bool)
		connectChan := make(chan bool, *numClients)
		handshakeChan := make(chan bool)
		connectLimiter := make(chan bool, *connectionLimit)
		for i := 0; i < totalClients; i++ {
			c := NewClient(serverList[i % len(serverList)], port, &Config{secure: *secure, delay: *delay})
			connectLimiter <- true
			go func(c *Client) {
				firstConnect := true
				firstHandshake := true
				for {
					err := c.Connect()
					if err != nil {
						time.Sleep(2 * time.Second)
					}

					if firstConnect {
						<-connectLimiter
						connectChan <- true
						firstConnect = false
					}

					<-waitHandshake
					err = c.Handshake()
					if err != nil {
						time.Sleep(2 * time.Second)
						continue
					}
					if firstHandshake {
						handshakeChan <- true
						firstHandshake = true
					}

					<-waitRegister
					c.Register()
					c.Run()
				}
			}(c)
		}

		for i := 0; i < totalClients; i++ {
			<-connectChan
		}
		close(waitHandshake)

		for i := 0; i < totalClients; i++ {
			<-handshakeChan
		}
		close(waitRegister)
	}()

	var tmp Metrics
	for {
		for k, v := range metrics {
			fmt.Printf("%s: %d ", k, v)
			tmpV, ok := tmp[k]
			if ok {
				fmt.Printf("(%d) ", v-tmpV)
			}
		}
		fmt.Println()
		tmp = metrics.Copy()
		time.Sleep(1 * time.Second)
	}
}
