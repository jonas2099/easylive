package stream

import (
	"github.com/haroldleong/easylive/cache"
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/container"
	"github.com/haroldleong/easylive/util"
	"github.com/orcaman/concurrent-map"
	log "github.com/sirupsen/logrus"
	"time"
)

type AppStream struct {
	anchorStream    *Stream            // 主播流
	audienceStreams cmap.ConcurrentMap // 观众流 map[string]*Stream
	cache           *cache.Cache       // 流的cache
}

func NewAppStream() *AppStream {
	return &AppStream{
		audienceStreams: cmap.New(),
		cache:           cache.NewCache(),
	}
}

func (as *AppStream) ReadingData(conn *conn.Conn) {
	startTime := time.Now()
	defer func() {
		log.Infof("ReadingData.stream end,take %d m", time.Since(startTime).Minutes())
		// TODO 销毁AppStream
	}()
	as.anchorStream = New(conn, true)
	for {
		cs := as.anchorStream.getStreamChunkStream()
		if cs == nil {
			return
		}
		p := container.GetPacketByChunk(cs)
		as.cache.Write(p)
		log.Debugf("ReadingData.send %d ", as.audienceStreams.Count())
		for _, tmp := range as.audienceStreams.Items() {
			audienceStream := tmp.(*Stream)
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
	stream := New(conn, false)
	as.audienceStreams.Set(stream.id, stream)
	go as.CheckPull(stream)
	go stream.writeToAudience()
	log.Debugf("AddAudienceWriteEvent success")
	return nil
}

func (as *AppStream) CheckPull(s *Stream) {
	for {
		if cs := GetChunk(s.conn); cs == nil {
			// 移除stream
			as.audienceStreams.Remove(s.id)
			log.Errorf("CheckPull.close")
			break
		} else {
			log.Debugf("CheckPull.cs:%v", util.JSON(cs))
		}
	}
}
