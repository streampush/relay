package main

import (
	"github.com/nareix/joy4/av/pubsub"
	"github.com/nareix/joy4/format/rtmp"
)

type Endpoint struct {
	Name string `json:"name"`
	URL  string `json:"url"`

	Dest      *rtmp.Conn
	Connected bool
	ConnErr   error
}

func (self *Endpoint) Update(new Endpoint) {
	self.Name = new.Name
	self.URL = new.URL
}

type Restream struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Endpoints map[string]*Endpoint `json:"endpoints"`
	Channel   chan string
	Origin    *rtmp.Conn
	Queue     *pubsub.Queue
	Streaming bool
}
