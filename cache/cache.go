package cache

import (
	"github.com/gwuhaolin/livego/av"
	"github.com/haroldleong/easylive/entity"
)

type Cache struct {
	gop      *GopCache
	videoSeq *SpecialCache
	audioSeq *SpecialCache
	metadata *SpecialCache
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(1),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

func (cache *Cache) Write(p *entity.Packet) {
	if p.IsMetadata {
		cache.metadata.Write(p)
		return
	} else {
		if !p.IsVideo {
			ah, ok := p.Header.(av.AudioPacketHeader)
			if ok {
				if ah.SoundFormat() == av.SOUND_AAC &&
					ah.AACPacketType() == av.AAC_SEQHDR {
					cache.audioSeq.Write(p)
					return
				} else {
					return
				}
			}
		} else {
			vh, ok := p.Header.(av.VideoPacketHeader)
			if ok {
				if vh.IsSeq() {
					cache.videoSeq.Write(p)
					return
				}
			} else {
				return
			}
		}
	}
	cache.gop.Write(p)
}

func (cache *Cache) Send(pChan chan *entity.Packet) error {
	if err := cache.metadata.Send(pChan); err != nil {
		return err
	}

	if err := cache.videoSeq.Send(pChan); err != nil {
		return err
	}

	if err := cache.audioSeq.Send(pChan); err != nil {
		return err
	}

	if err := cache.gop.Send(pChan); err != nil {
		return err
	}

	return nil
}
