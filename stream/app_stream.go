package stream

import (
	"github.com/haroldleong/easylive/cache"
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/container"
	log "github.com/sirupsen/logrus"
)

type AppStream struct {
	anchorStream    *Stream
	audienceStreams []*Stream
	cache           *cache.Cache
}

func NewAppStream() *AppStream {
	return &AppStream{
		audienceStreams: []*Stream{},
		cache:           cache.NewCache(),
	}
}

func (as *AppStream) ReadingData(conn *conn.Conn) {
	as.anchorStream = &Stream{conn: conn}
	for {
		cs := as.anchorStream.getStreamChunkStream()
		if cs == nil {
			return
		}
		p := container.GetPacketByChunk(cs)
		as.cache.Write(p)
		for _, audienceStream := range as.audienceStreams {
			// H264的码流结构主要由SPS、 PPS、 IDR 帧（包含一个或多个 I-Slice）、 P 帧（包含一个或多个P-Slice）、 B 帧（包含一个或多个 B-Slice）等部分组成
			if !audienceStream.init {
				if err := as.cache.Send(audienceStream.packetQueue); err != nil {
					log.Errorf("ReadingData.Send.%v", err)
				}
				audienceStream.init = true
			} else {
				audienceStream.packetQueue <- p
			}
		}
	}
}

func (as *AppStream) AddAudienceWriteEvent(conn *conn.Conn) error {
	stream := New(conn)
	as.audienceStreams = append(as.audienceStreams, stream)
	go stream.CheckPull()
	go stream.writeToAudience()
	log.Debugf("AddAudienceWriteEvent success")
	return nil
}
