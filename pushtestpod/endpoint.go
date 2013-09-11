package main

import (
	"github.com/oremj/go-simplepush-client/pushclient"
	"time"
)

type endPoint struct {
	reg       *pushclient.RegisterResponse
	version   int
	done      chan bool
	delay     time.Duration
	delayChan <-chan time.Time
}

func NewEndpoint(reg *pushclient.RegisterResponse, delay int) *endPoint {
	e := &endPoint{
		reg:     reg,
		version: 1,
		done:    make(chan bool),
		delay:   time.Duration(delay) * time.Second,
	}

	e.setTimer()
	return e
}

func (e *endPoint) setTimer() {
	e.delayChan = time.After(e.delay)
}

func (e *endPoint) sendPing() (err error) {
	<-e.delayChan
	defer e.setTimer()
	err = SendPing(e.reg.PushEndpoint, e.version)
	if err != nil {
		counterChan <- &stat{"put_fail", 1}
	} else {
		counterChan <- &stat{"put_ok", 1}
		timeout := time.After(15 * time.Second)
		select {
		case <-timeout:
			e.done <- true
		case <-e.done:
		}
	}
	return
}
