package main

import (
	"github.com/oremj/go-simplepush-client/pushclient"
	"math/rand"
	"time"
)

type endPoint struct {
	reg     *pushclient.RegisterResponse
	version int
	notify chan bool
}

func NewEndpoint(reg *pushclient.RegisterResponse) *endPoint {
	e := &endPoint{
		reg:     reg,
		version: 1,
		notify: make(chan bool, 1),
	}

	return e
}

func (e *endPoint) run(delay int) (err error) {
	for {
		err = e.sendPing()
		if err != nil {
			return
		}
		select {
		case <-time.After(10 * time.Second):
			incStat("update_timeout")
			close(e.notify)
			return
		case <-e.notify:
			incStat("update_ok")
			e.version++
		}

		range_ := delay / 2
		durDelay := rand.Intn(range_*2) + (delay - range_)
		time.Sleep(time.Duration(durDelay) * time.Millisecond)
	}
}

func (e *endPoint) sendPing() (err error) {
	err = SendPing(e.reg.PushEndpoint, e.version)
	if err != nil {
		incStat("put_fail")
		return
	}
	incStat("put_ok")
	return
}
