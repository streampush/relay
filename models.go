package main

import (
	"log"
	"time"

	"github.com/nareix/joy4/av/pubsub"
	"github.com/nareix/joy4/format/rtmp"
)

type ConnStats struct {
	TxBytes uint64 `json:"txBytes"`
	RxBytes uint64 `json:"rxBytes"`

	Bitrate float64 `json:"bitrate"`
}

type EventMsg struct {
	Level     string `json:"level"`
	Message   string `json:"msg"`
	Timestamp int64  `json:"time"`
}

type Endpoint struct {
	Name string `json:"name"`
	URL  string `json:"url"`

	Dest      *rtmp.Conn `json:"-"`
	Connected bool       `json:"connected"`
	ConnErr   error      `json:"connectErr,omitempty"`
	Stats     ConnStats  `json:"stats"`
}

func (self *Endpoint) Update(new Endpoint) {
	self.Name = new.Name
	self.URL = new.URL
}

type Restream struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Endpoints map[string]*Endpoint `json:"endpoints"`
	Channel   chan string          `json:"-"`
	Origin    *rtmp.Conn           `json:"-"`
	Queue     *pubsub.Queue        `json:"-"`
	Streaming bool                 `json:"streaming"`
	Stats     ConnStats            `json:"stats"`
	Events    []EventMsg           `json:"events"`
}

func (self *Restream) AddEvent(level, message string) {
	log.Printf("%s][%s] %s", self.Name, level, message)
	self.Events = append(self.Events, EventMsg{
		Level:     level,
		Message:   message,
		Timestamp: time.Now().UnixNano(),
	})
}
