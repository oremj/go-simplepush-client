package main

import (
	"github.com/oremj/go-simplepush-client/pushclient"
	"time"
)

type endPoint struct {
	reg     *pushclient.RegisterResponse
	version int
	done    chan bool
}

func NewEndpoint(reg *pushclient.RegisterResponse) *endPoint {
	return &endPoint{reg: reg, version: 1, done: make(chan bool)}
}

func (e *endPoint) sendPing() (err error) {
	err = SendPing(e.reg.PushEndpoint, e.version)
	if err != nil {
		counterChan <- &stat{"put_fail", 1}
	} else {
		go func(e *endPoint) {
			timeout := time.After(5 * time.Second)
			select {
			case <-timeout:
				counterChan <- &stat{"update_timeout", 1}
			case <-e.done:
				counterChan <- &stat{"update_ok", 1}
			}
		}(e)
		counterChan <- &stat{"put_ok", 1}
	}
	return
}
