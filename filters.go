package main

import (
	"log"
	"time"

	"github.com/nareix/joy4/av"
)

// CalcBitrate calculates bitrate and stores it in a Restream
type CalcBitrate struct {
	Bitrate float64

	bitCnt   int64
	lastTime int64

	Restream *Restream
	Endpoint *Endpoint
}

func (self *CalcBitrate) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	// Sample abt. every half second
	if self.lastTime+(5e8) < time.Now().UnixNano() {
		timeDelta := ((float64(time.Now().UnixNano()) - float64(self.lastTime)) * 1e-9)
		self.Bitrate = float64(self.bitCnt/100) / timeDelta

		if self.Restream != nil {
			self.Restream.Stats.Bitrate = self.Bitrate
		}

		if self.Endpoint != nil {
			log.Fatal("Endpoint bitrate calculation not yet implemented")
			self.Endpoint.Stats.Bitrate = self.Bitrate
		}

		log.Printf("%s bitrate %.3f kb/s\n", self.Restream.Name, self.Bitrate)

		self.lastTime = time.Now().UnixNano()
		self.bitCnt = 0
	}

	// TODO: Pretty sure this is somewhat incorrect
	// values seem to be off by about 700 kb/s
	self.bitCnt += int64(len(pkt.Data))

	drop = false
	return
}
