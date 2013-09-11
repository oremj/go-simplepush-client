package main

import (
	"github.com/oremj/go-simplepush-client/pushclient"
	"time"
)

type endPoint struct {
	reg     *pushclient.RegisterResponse
	version int
	success chan bool
}

func NewEndpoint(reg *pushclient.RegisterResponse) *endPoint {
	e := &endPoint{
		reg:     reg,
		version: 1,
		success: make(chan bool),
	}

	return e
}

func (e *endPoint) sendPing() (err error) {
	err = SendPing(e.reg.PushEndpoint, e.version)
	if err != nil {
		incStat("put_fail")
		return
	}
	incStat("put_ok")
	select {
	case <-time.After(10 * time.Second):
		incStat("update_timeout")
		close(e.success)
	case e.success <- true:
		incStat("update_ok")
	}
	return
}
