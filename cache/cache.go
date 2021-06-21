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
		if p.IsAudio {
			ah, ok := p.Header.(av.AudioPacketHeader)
			if ok {
				if ah.SoundFormat() == av.SOUND_AAC &&
					ah.AACPacketType() == av.AAC_SEQHDR {
					// 必须要在发送第一个 AAC raw 包之前发送这个 AAC sequence header 包
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
					// 在给AVC解码器送数据流之前一定要把sps和pps信息送出否则的话解码器不能正常解码
					// SPS即Sequence Paramater Set，又称作序列参数集,作为全局参数
					// Picture Paramater Set(PPS)
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
