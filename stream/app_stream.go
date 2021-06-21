package stream

import (
	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/container/flv"
	"github.com/haroldleong/easylive/cache"
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/entity"
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

func DemuxH(p *entity.Packet) error {
	var tag flv.Tag
	_, err := tag.ParseMediaTagHeader(p.CStream.Data, p.IsVideo)
	if err != nil {
		return err
	}
	p.Header = &tag

	return nil
}

func (as *AppStream) ReadingData(conn *conn.Conn) {
	as.anchorStream = &Stream{conn: conn}
	for {
		cs := as.anchorStream.getStreamChunkStream()
		p := &entity.Packet{CStream: *cs}
		p.IsMetadata = cs.TypeID == av.TAG_SCRIPTDATAAMF0 || cs.TypeID == av.TAG_SCRIPTDATAAMF3
		if p.IsMetadata {
			log.Errorf("")
		}
		p.IsAudio = cs.TypeID == av.TAG_AUDIO
		p.IsVideo = cs.TypeID == av.TAG_VIDEO
		if err := DemuxH(p); err != nil {
			log.Errorf("ReadingData.DemuxH.%v", err)
		}
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
