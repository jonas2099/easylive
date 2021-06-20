package stream

import (
	"github.com/haroldleong/easylive/conn"
	log "github.com/sirupsen/logrus"
)

type AppStream struct {
	anchorStream    *Stream
	audienceStreams []*Stream
}

func NewAppStream() *AppStream {
	return &AppStream{
		audienceStreams: []*Stream{},
	}
}

func (as *AppStream) ReadingData(conn *conn.Conn) {
	as.anchorStream = &Stream{conn: conn}
	for {
		cs := as.anchorStream.getStreamChunkStream()
		for _, audienceStream := range as.audienceStreams {
			audienceStream.packetQueue <- cs
		}
	}
}

func (as *AppStream) AddAudienceWriteEvent(conn *conn.Conn) error {
	stream := New(conn)
	as.audienceStreams = append(as.audienceStreams, stream)
	go stream.writeToAudience()
	log.Debugf("AddAudienceWriteEvent success")
	return nil
}
