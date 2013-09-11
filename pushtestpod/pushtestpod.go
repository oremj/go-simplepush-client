package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/oremj/go-simplepush-client/pushclient"
	"net/http"
	"strings"
	"time"
)

var server = flag.String("server", "localhost", "Pushgo Server Name")
var secure = flag.Bool("secure", false, "Use wss/https")
var numClients = flag.Int("clients", 1, "Number of concurrent clients")

type endPoint struct {
	reg     *pushclient.RegisterResponse
	version int
	done chan bool
	timeout <-chan time.Time
}

func (e *endPoint) sendPing() (err error) {
	e.timeout = time.After(5 * time.Second)
	go func(e *endPoint) {
		select {
		case <-e.timeout:
			statChan <- &stat{"ping_timeout", 1}
		case <-e.done:
			statChan <- &stat{"ping_recv", 1}
		}
	}(e)
	err = SendPing(e.reg.PushEndpoint, e.version)
	if err != nil {
		statChan <- &stat{"put_fail", 1}
	} else {
		statChan <- &stat{"ping_sent", 1}
	}
	return
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
		statChan <- &stat{"conn_fail", 1}
		return
	}
	statChan <- &stat{"conn_ok", 1}
	go pc.Run()
	<-waitConnect
	pc.Register()

	endPoints := make(map[string]*endPoint)
	for {
		select {
		case reg := <-c.Register:
			e := &endPoint{reg, 1, make(chan bool), time.After(3 * time.Millisecond)}
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

var statChan chan *stat
var waitConnect chan bool

func main() {
	flag.Parse()
	statChan = make(chan *stat, 100000)
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
		go RunClient(*server, port, *secure, c)
		time.Sleep(100 * time.Millisecond)
	}
	connected := false
	for {
		s := <-statChan
		metrics[s.key] += s.val
		if !connected && metrics["conn_ok"] + metrics["conn_fail"] >= *numClients {
			for i := 0; i < *numClients; i++ {
				waitConnect <- true
			}
			connected = true
		}
		fmt.Println(metrics)
	}
}
