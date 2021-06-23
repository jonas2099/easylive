package cache

import (
	"github.com/haroldleong/easylive/container"
)

type Cache struct {
	gop      *GopCache
	videoSeq *PackageCache
	audioSeq *PackageCache
	metadata *PackageCache
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(2),
		videoSeq: NewPackageCache(),
		audioSeq: NewPackageCache(),
		metadata: NewPackageCache(),
	}
}

func (cache *Cache) Write(p *container.Packet) {
	if p.IsType(container.PacketTypeMetaData) {
		cache.metadata.Write(p)
		return
	} else if p.IsType(container.PacketTypeAudio) {
		if p.Tag.AudioHeader.SoundFormat == container.SoundAAC &&
			p.Tag.AudioHeader.AACPacketType == container.AACSequenceHeader {
			// 必须要在发送第一个 AAC raw 包之前发送这个 AAC sequence header 包
			cache.audioSeq.Write(p)
			return
		}
	} else {
		if p.Tag.VideoHeader.FrameType == container.FrameKey &&
			p.Tag.VideoHeader.AVCPacketType == container.AVCSequenceHeader {
			// 在给AVC解码器送数据流之前一定要把sps和pps信息送出否则的话解码器不能正常解码
			// 序列参数集 Sequence Parameter Set，作为全局参数
			// 图像参数集 Picture Parameter Set
			cache.videoSeq.Write(p)
			return
		}
	}
	cache.gop.Write(p)
}

func (cache *Cache) Send(pChan chan *container.Packet) error {
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
