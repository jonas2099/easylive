package cache

import (
	"bytes"
	"github.com/haroldleong/easylive/entity"

	"github.com/gwuhaolin/livego/protocol/amf"

	log "github.com/sirupsen/logrus"
)

const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}
	if _, err := encoder.Encode(b, SetDataFrame, amf.AMF0); err != nil {
		log.Fatal(err)
	}
	setFrameFrame = b.Bytes()
}

type SpecialCache struct {
	full bool
	p    *entity.Packet
}

func NewSpecialCache() *SpecialCache {
	return &SpecialCache{}
}

func (specialCache *SpecialCache) Write(p *entity.Packet) {
	specialCache.p = p
	specialCache.full = true
}

func (specialCache *SpecialCache) Send(pChan chan *entity.Packet) error {
	if !specialCache.full {
		return nil
	}
	pChan <- specialCache.p
	return nil
}