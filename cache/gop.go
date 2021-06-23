package cache

import (
	"fmt"
	"github.com/haroldleong/easylive/container"
)

var (
	maxGOPCapacity = 1024
	ErrGopTooBig   = fmt.Errorf("gop size over limit")
)

type gop struct {
	packets []*container.Packet
}

func newArray() *gop {
	ret := &gop{
		packets: make([]*container.Packet, 0, maxGOPCapacity),
	}
	return ret
}

func (g *gop) reset() {
	g.packets = g.packets[:0]
}

func (g *gop) write(packet *container.Packet) error {
	if len(g.packets) >= maxGOPCapacity {
		return ErrGopTooBig
	}
	g.packets = append(g.packets, packet)
	return nil
}

func (g *gop) readAndSend(pChan chan *container.Packet) error {
	var err error
	for _, packet := range g.packets {
		pChan <- packet
	}
	return err
}

type GopCache struct {
	start     bool
	num       int
	count     int
	nextindex int
	gops      []*gop
}

func NewGopCache(num int) *GopCache {
	return &GopCache{
		count: num,
		gops:  make([]*gop, num),
	}
}

func (gopCache *GopCache) writeToArray(chunk *container.Packet, startNew bool) error {
	var ginc *gop
	if startNew {
		ginc = gopCache.gops[gopCache.nextindex]
		if ginc == nil {
			ginc = newArray()
			gopCache.num++
			gopCache.gops[gopCache.nextindex] = ginc
		} else {
			ginc.reset()
		}
		gopCache.nextindex = (gopCache.nextindex + 1) % gopCache.count
	} else {
		ginc = gopCache.gops[(gopCache.nextindex+1)%gopCache.count]
	}
	ginc.write(chunk)

	return nil
}

func (gopCache *GopCache) Write(p *container.Packet) {
	var ok bool
	if p.IsType(container.PacketTypeVideo) {
		if p.Tag.IsKeyFrame() && !p.Tag.IsSeq() {
			ok = true
		}
	}
	if ok || gopCache.start {
		gopCache.start = true
		_ = gopCache.writeToArray(p, ok)
	}
}

func (gopCache *GopCache) sendTo(pChan chan *container.Packet) error {
	var err error
	pos := (gopCache.nextindex + 1) % gopCache.count
	for i := 0; i < gopCache.num; i++ {
		index := (pos - gopCache.num + 1) + i
		if index < 0 {
			index += gopCache.count
		}
		g := gopCache.gops[index]
		err = g.readAndSend(pChan)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gopCache *GopCache) Send(pChan chan *container.Packet) error {
	return gopCache.sendTo(pChan)
}
