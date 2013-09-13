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

	metrics := make(Metrics)
	go func() {
		for s := range counterChan {
			metrics[s.key] += s.val
		}
	}()

	port := 80
	if *secure {
		port = 443
	}

	go func() {
		clients := make([]*Client, 0, *numClients)
		connectionLimiter := make(chan bool, 300)
		for i := 0; i < *numClients; i++ {
			connectionLimiter <- true
			c := NewClient(*server, port, &Config{secure: *secure, delay: *delay})
			clients = append(clients, c)
			go func(c *Client) {
				for {
					err := c.Connect()
					if err != nil {
						time.Sleep(2 * time.Second)
						continue
					}
					c.SendReg()

					<-connectionLimiter

					err = c.Run()
					if err != nil {
						time.Sleep(2 * time.Second)
					}
				}
			}(c)
		}
	}()

	var tmp Metrics
	for {
		for k, v := range metrics {
			fmt.Printf("%s: %d ", k, v)
			tmpV, ok := tmp[k]
			if ok {
				fmt.Printf("(%d) ", v - tmpV)
			}
		}
		fmt.Println()
		tmp = metrics.Copy()
		time.Sleep(1 * time.Second)
	}
}
